#!/usr/bin/env node

/**
 * WebSocket Test Script for Plivo PubSub Gateway
 * 
 * This script tests WebSocket functionality including:
 * - Connection with JWT token
 * - Subscribe/Unsubscribe operations
 * - Message publishing
 * - Ping/Pong
 * 
 * Usage:
 * 1. Start the service: JWT_SECRET_KEY="test-secret" go run main.go
 * 2. Register a user and get a token
 * 3. Run: node test-websocket.js <token>
 */

const WebSocket = require('ws');

// Configuration
const WS_URL = 'ws://localhost:8000/ws';
const TOPIC_NAME = 'test-topic';

// Colors for console output
const colors = {
    reset: '\x1b[0m',
    red: '\x1b[31m',
    green: '\x1b[32m',
    yellow: '\x1b[33m',
    blue: '\x1b[34m',
    magenta: '\x1b[35m',
    cyan: '\x1b[36m'
};

function log(message, color = 'reset') {
    console.log(`${colors[color]}${message}${colors.reset}`);
}

function logError(message) {
    log(`❌ ${message}`, 'red');
}

function logSuccess(message) {
    log(`✅ ${message}`, 'green');
}

function logInfo(message) {
    log(`ℹ️  ${message}`, 'blue');
}

function logWarning(message) {
    log(`⚠️  ${message}`, 'yellow');
}

// Test results
let testsPassed = 0;
let testsFailed = 0;

function runTest(testName, testFn) {
    return new Promise((resolve) => {
        log(`\n🧪 Test: ${testName}`, 'cyan');
        testFn()
            .then(() => {
                logSuccess(`${testName} passed`);
                testsPassed++;
                resolve();
            })
            .catch((error) => {
                logError(`${testName} failed: ${error.message}`);
                testsFailed++;
                resolve();
            });
    });
}

async function testWebSocketConnection(token) {
    return new Promise((resolve, reject) => {
        const ws = new WebSocket(`${WS_URL}?token=${token}`);
        let connected = false;
        let messagesReceived = 0;
        let pongReceived = false;
        let subscriptionAck = false;
        let publishAck = false;
        let eventReceived = false;

        const timeout = setTimeout(() => {
            if (!connected) {
                ws.close();
                reject(new Error('Connection timeout'));
            }
        }, 5000);

        ws.on('open', () => {
            connected = true;
            clearTimeout(timeout);
            logInfo('WebSocket connected successfully');
            
            // Test 1: Send ping
            const pingMessage = {
                type: 'ping',
                request_id: 'test-ping-001'
            };
            ws.send(JSON.stringify(pingMessage));
            logInfo('Sent ping message');

            // Test 2: Subscribe to topic
            setTimeout(() => {
                const subscribeMessage = {
                    type: 'subscribe',
                    topic: TOPIC_NAME,
                    last_n: 5,
                    request_id: 'test-subscribe-001'
                };
                ws.send(JSON.stringify(subscribeMessage));
                logInfo('Sent subscribe message');
            }, 100);

            // Test 3: Publish a message
            setTimeout(() => {
                const publishMessage = {
                    type: 'publish',
                    topic: TOPIC_NAME,
                    message: {
                        id: 'test-msg-001',
                        payload: {
                            text: 'Hello from WebSocket test!',
                            timestamp: new Date().toISOString()
                        }
                    },
                    request_id: 'test-publish-001'
                };
                ws.send(JSON.stringify(publishMessage));
                logInfo('Sent publish message');
            }, 200);

            // Test 4: Unsubscribe
            setTimeout(() => {
                const unsubscribeMessage = {
                    type: 'unsubscribe',
                    topic: TOPIC_NAME,
                    request_id: 'test-unsubscribe-001'
                };
                ws.send(JSON.stringify(unsubscribeMessage));
                logInfo('Sent unsubscribe message');
            }, 300);

            // Close connection after tests
            setTimeout(() => {
                ws.close();
            }, 1000);
        });

        ws.on('message', (data) => {
            try {
                const message = JSON.parse(data.toString());
                messagesReceived++;
                logInfo(`Received message: ${JSON.stringify(message, null, 2)}`);

                // Check message types
                switch (message.type) {
                    case 'pong':
                        if (message.request_id === 'test-ping-001') {
                            pongReceived = true;
                            logSuccess('Pong received');
                        }
                        break;
                    case 'ack':
                        if (message.request_id === 'test-subscribe-001') {
                            subscriptionAck = true;
                            logSuccess('Subscription acknowledged');
                        } else if (message.request_id === 'test-publish-001') {
                            publishAck = true;
                            logSuccess('Publish acknowledged');
                        }
                        break;
                    case 'event':
                        if (message.topic === TOPIC_NAME) {
                            eventReceived = true;
                            logSuccess('Event message received');
                        }
                        break;
                    case 'error':
                        logError(`Error received: ${message.error?.message || 'Unknown error'}`);
                        break;
                }
            } catch (error) {
                logError(`Failed to parse message: ${error.message}`);
            }
        });

        ws.on('close', () => {
            logInfo('WebSocket connection closed');
            
            // Validate test results
            if (pongReceived && subscriptionAck && publishAck && eventReceived) {
                resolve();
            } else {
                reject(new Error(`Missing responses: pong=${pongReceived}, subAck=${subscriptionAck}, pubAck=${publishAck}, event=${eventReceived}`));
            }
        });

        ws.on('error', (error) => {
            clearTimeout(timeout);
            reject(error);
        });
    });
}

async function testWebSocketWithoutToken() {
    return new Promise((resolve, reject) => {
        const ws = new WebSocket(WS_URL);
        
        const timeout = setTimeout(() => {
            ws.close();
            reject(new Error('Connection should have been rejected'));
        }, 3000);

        ws.on('open', () => {
            clearTimeout(timeout);
            ws.close();
            reject(new Error('Connection should have been rejected without token'));
        });

        ws.on('error', (error) => {
            clearTimeout(timeout);
            if (error.message.includes('401') || error.message.includes('Unauthorized')) {
                resolve();
            } else {
                reject(new Error(`Unexpected error: ${error.message}`));
            }
        });
    });
}

async function testWebSocketWithInvalidToken() {
    return new Promise((resolve, reject) => {
        const ws = new WebSocket(`${WS_URL}?token=invalid-token`);
        
        const timeout = setTimeout(() => {
            ws.close();
            reject(new Error('Connection should have been rejected'));
        }, 3000);

        ws.on('open', () => {
            clearTimeout(timeout);
            ws.close();
            reject(new Error('Connection should have been rejected with invalid token'));
        });

        ws.on('error', (error) => {
            clearTimeout(timeout);
            if (error.message.includes('401') || error.message.includes('Unauthorized')) {
                resolve();
            } else {
                reject(new Error(`Unexpected error: ${error.message}`));
            }
        });
    });
}

async function main() {
    const token = process.argv[2];
    
    if (!token) {
        logError('Usage: node test-websocket.js <jwt-token>');
        logInfo('To get a token:');
        logInfo('1. Start the service: JWT_SECRET_KEY="test-secret" go run main.go');
        logInfo('2. Register a user: curl -X POST http://localhost:8000/users/register -H "Content-Type: application/json" -d \'{"username": "testuser", "password": "password123"}\'');
        logInfo('3. Login: curl -X POST http://localhost:8000/users/login -H "Content-Type: application/json" -d \'{"username": "testuser", "password": "password123"}\'');
        process.exit(1);
    }

    log('🚀 Starting WebSocket Tests for Plivo PubSub Gateway', 'magenta');
    log(`Token: ${token.substring(0, 50)}...`, 'blue');

    try {
        // Test 1: WebSocket without token (should fail)
        await runTest('WebSocket without token', testWebSocketWithoutToken);

        // Test 2: WebSocket with invalid token (should fail)
        await runTest('WebSocket with invalid token', testWebSocketWithInvalidToken);

        // Test 3: WebSocket with valid token and full functionality
        await runTest('WebSocket with valid token and full functionality', () => testWebSocketConnection(token));

    } catch (error) {
        logError(`Test suite failed: ${error.message}`);
    }

    // Print results
    log('\n📊 Test Results Summary', 'cyan');
    log('========================', 'cyan');
    logSuccess(`Tests Passed: ${testsPassed}`);
    if (testsFailed > 0) {
        logError(`Tests Failed: ${testsFailed}`);
    }
    log(`Total Tests: ${testsPassed + testsFailed}`);

    if (testsFailed === 0) {
        log('\n🎉 All WebSocket tests passed!', 'green');
    } else {
        log('\n⚠️  Some tests failed. Check the service logs.', 'yellow');
        process.exit(1);
    }
}

// Handle uncaught exceptions
process.on('uncaughtException', (error) => {
    logError(`Uncaught exception: ${error.message}`);
    process.exit(1);
});

process.on('unhandledRejection', (reason, promise) => {
    logError(`Unhandled rejection at: ${promise}, reason: ${reason}`);
    process.exit(1);
});

main().catch((error) => {
    logError(`Main function failed: ${error.message}`);
    process.exit(1);
});
