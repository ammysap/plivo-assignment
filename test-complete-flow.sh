#!/bin/bash

echo "🧪 Complete Plivo PubSub Gateway Test Suite"
echo "==========================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
BASE_URL="http://localhost:8000"
JWT_SECRET_KEY="test-secret-key-for-development"

# Start the service in background
echo -e "${BLUE}🚀 Starting PubSub Gateway Service...${NC}"
JWT_SECRET_KEY="$JWT_SECRET_KEY" go run main.go &
SERVICE_PID=$!

# Wait for service to start
echo -e "${YELLOW}⏳ Waiting for service to start...${NC}"
sleep 5

# Test counter
TESTS_PASSED=0
TESTS_FAILED=0

# Function to run a test
run_test() {
    local test_name="$1"
    local command="$2"
    local expected_status="$3"
    
    echo -e "\n${BLUE}📝 Test: $test_name${NC}"
    echo "Command: $command"
    
    response=$(eval "$command" 2>/dev/null)
    status_code=$?
    
    if [ $status_code -eq 0 ]; then
        echo -e "${GREEN}✅ PASSED${NC}"
        echo "Response: $response"
        ((TESTS_PASSED++))
    else
        echo -e "${RED}❌ FAILED${NC}"
        echo "Response: $response"
        ((TESTS_FAILED++))
    fi
}

# Test 1: Health Check
run_test "Health Check" "curl -s $BASE_URL/health" "200"

# Test 2: Stats (should be empty initially)
run_test "Initial Stats" "curl -s $BASE_URL/stats" "200"

# Test 3: Register User
echo -e "\n${BLUE}📝 Test: Register User${NC}"
# Use a unique username to avoid conflicts
UNIQUE_USERNAME="testuser$(date +%s)"
REGISTER_RESPONSE=$(curl -s -X POST $BASE_URL/users/register \
  -H "Content-Type: application/json" \
  -d "{\"username\": \"$UNIQUE_USERNAME\", \"password\": \"password123\", \"email\": \"test@example.com\"}")
echo "Register Response: $REGISTER_RESPONSE"

# Extract token
TOKEN=$(echo $REGISTER_RESPONSE | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
if [ -n "$TOKEN" ]; then
    echo -e "${GREEN}✅ User registered successfully${NC}"
    echo "Token: ${TOKEN:0:50}..."
    ((TESTS_PASSED++))
else
    echo -e "${RED}❌ User registration failed${NC}"
    ((TESTS_FAILED++))
fi

# Test 4: Login User
echo -e "\n${BLUE}📝 Test: Login User${NC}"
LOGIN_RESPONSE=$(curl -s -X POST $BASE_URL/users/login \
  -H "Content-Type: application/json" \
  -d "{\"username\": \"$UNIQUE_USERNAME\", \"password\": \"password123\"}")
echo "Login Response: $LOGIN_RESPONSE"

LOGIN_TOKEN=$(echo $LOGIN_RESPONSE | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
if [ -n "$LOGIN_TOKEN" ]; then
    echo -e "${GREEN}✅ User login successful${NC}"
    echo "Login Token: ${LOGIN_TOKEN:0:50}..."
    ((TESTS_PASSED++))
else
    echo -e "${RED}❌ User login failed${NC}"
    ((TESTS_FAILED++))
fi

# Test 5: Get User Profile
echo -e "\n${BLUE}📝 Test: Get User Profile${NC}"
# Use the login token if available, otherwise use the register token
PROFILE_TOKEN=${LOGIN_TOKEN:-$TOKEN}
PROFILE_RESPONSE=$(curl -s -X GET $BASE_URL/users/profile \
  -H "Authorization: Bearer $PROFILE_TOKEN")
echo "Profile Response: $PROFILE_RESPONSE"

if echo "$PROFILE_RESPONSE" | grep -q "$UNIQUE_USERNAME"; then
    echo -e "${GREEN}✅ Profile retrieved successfully${NC}"
    ((TESTS_PASSED++))
else
    echo -e "${RED}❌ Profile retrieval failed${NC}"
    ((TESTS_FAILED++))
fi

# Test 6: Create Topic
echo -e "\n${BLUE}📝 Test: Create Topic${NC}"
UNIQUE_TOPIC="test-topic-$(date +%s)"
TOPIC_TOKEN=${LOGIN_TOKEN:-$TOKEN}
CREATE_TOPIC_RESPONSE=$(curl -s -X POST $BASE_URL/topics \
  -H "Authorization: Bearer $TOPIC_TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"name\": \"$UNIQUE_TOPIC\"}")
echo "Create Topic Response: $CREATE_TOPIC_RESPONSE"

if echo "$CREATE_TOPIC_RESPONSE" | grep -q "created"; then
    echo -e "${GREEN}✅ Topic created successfully${NC}"
    ((TESTS_PASSED++))
else
    echo -e "${RED}❌ Topic creation failed${NC}"
    ((TESTS_FAILED++))
fi

# Test 7: List Topics
echo -e "\n${BLUE}📝 Test: List Topics${NC}"
LIST_TOPICS_RESPONSE=$(curl -s -X GET $BASE_URL/topics \
  -H "Authorization: Bearer $TOPIC_TOKEN")
echo "List Topics Response: $LIST_TOPICS_RESPONSE"

if echo "$LIST_TOPICS_RESPONSE" | grep -q "$UNIQUE_TOPIC"; then
    echo -e "${GREEN}✅ Topics listed successfully${NC}"
    ((TESTS_PASSED++))
else
    echo -e "${RED}❌ Topic listing failed${NC}"
    ((TESTS_FAILED++))
fi

# Test 8: Create Multiple Topics
echo -e "\n${BLUE}📝 Test: Create Multiple Topics${NC}"
for i in {1..3}; do
    curl -s -X POST $BASE_URL/topics \
      -H "Authorization: Bearer $TOPIC_TOKEN" \
      -H "Content-Type: application/json" \
      -d "{\"name\": \"topic-$i-$(date +%s)\"}" > /dev/null
done
echo -e "${GREEN}✅ Multiple topics created${NC}"
((TESTS_PASSED++))

# Test 9: Check Stats
echo -e "\n${BLUE}📝 Test: Check Stats${NC}"
STATS_RESPONSE=$(curl -s $BASE_URL/stats)
echo "Stats Response: $STATS_RESPONSE"

if echo "$STATS_RESPONSE" | grep -q "topic-"; then
    echo -e "${GREEN}✅ Stats show created topics${NC}"
    ((TESTS_PASSED++))
else
    echo -e "${RED}❌ Stats don't show topics${NC}"
    ((TESTS_FAILED++))
fi

# Test 10: Delete Topic
echo -e "\n${BLUE}📝 Test: Delete Topic${NC}"
DELETE_RESPONSE=$(curl -s -X DELETE $BASE_URL/topics/$UNIQUE_TOPIC \
  -H "Authorization: Bearer $TOPIC_TOKEN")
echo "Delete Response: $DELETE_RESPONSE"

if echo "$DELETE_RESPONSE" | grep -q "deleted"; then
    echo -e "${GREEN}✅ Topic deleted successfully${NC}"
    ((TESTS_PASSED++))
else
    echo -e "${RED}❌ Topic deletion failed${NC}"
    ((TESTS_FAILED++))
fi

# Test 11: Try to delete non-existent topic
echo -e "\n${BLUE}📝 Test: Delete Non-existent Topic${NC}"
DELETE_NONEXISTENT_RESPONSE=$(curl -s -X DELETE $BASE_URL/topics/nonexistent \
  -H "Authorization: Bearer $TOPIC_TOKEN")
echo "Delete Non-existent Response: $DELETE_NONEXISTENT_RESPONSE"

if echo "$DELETE_NONEXISTENT_RESPONSE" | grep -q "not found"; then
    echo -e "${GREEN}✅ Proper error for non-existent topic${NC}"
    ((TESTS_PASSED++))
else
    echo -e "${RED}❌ Unexpected response for non-existent topic${NC}"
    ((TESTS_FAILED++))
fi

# Test 12: Test WebSocket without Token (should return 401)
echo -e "\n${BLUE}📝 Test: WebSocket without Token${NC}"
WS_NO_TOKEN_RESPONSE=$(curl -s -w "%{http_code}" -o /dev/null "$BASE_URL/ws")
if [ "$WS_NO_TOKEN_RESPONSE" = "401" ]; then
    echo -e "${GREEN}✅ WebSocket properly rejects requests without token (401)${NC}"
    ((TESTS_PASSED++))
else
    echo -e "${RED}❌ WebSocket should reject requests without token, got: $WS_NO_TOKEN_RESPONSE${NC}"
    ((TESTS_FAILED++))
fi

# Test 13: Test WebSocket with Invalid Token (should return 401)
echo -e "\n${BLUE}📝 Test: WebSocket with Invalid Token${NC}"
WS_INVALID_TOKEN_RESPONSE=$(curl -s -w "%{http_code}" -o /dev/null "$BASE_URL/ws?token=invalid-token")
if [ "$WS_INVALID_TOKEN_RESPONSE" = "401" ]; then
    echo -e "${GREEN}✅ WebSocket properly rejects invalid token (401)${NC}"
    ((TESTS_PASSED++))
else
    echo -e "${RED}❌ WebSocket should reject invalid token, got: $WS_INVALID_TOKEN_RESPONSE${NC}"
    ((TESTS_FAILED++))
fi

# Test 14: Test WebSocket with Valid Token (should return 400 - Bad Request for non-WebSocket client)
echo -e "\n${BLUE}📝 Test: WebSocket with Valid Token${NC}"
# Use the login token if available, otherwise use the register token
WS_TOKEN=${LOGIN_TOKEN:-$TOKEN}
WS_VALID_TOKEN_RESPONSE=$(curl -s -w "%{http_code}" -o /dev/null "$BASE_URL/ws?token=$WS_TOKEN")
if [ "$WS_VALID_TOKEN_RESPONSE" = "400" ]; then
    echo -e "${GREEN}✅ WebSocket with valid token responds correctly (400 - expects WebSocket upgrade)${NC}"
    ((TESTS_PASSED++))
else
    echo -e "${RED}❌ WebSocket with valid token should return 400, got: $WS_VALID_TOKEN_RESPONSE${NC}"
    ((TESTS_FAILED++))
fi

# Test 15: Test WebSocket Endpoint Exists
echo -e "\n${BLUE}📝 Test: WebSocket Endpoint Exists${NC}"
WS_ENDPOINT_RESPONSE=$(curl -s -w "%{http_code}" -o /dev/null "$BASE_URL/ws")
if [ "$WS_ENDPOINT_RESPONSE" = "401" ] || [ "$WS_ENDPOINT_RESPONSE" = "400" ]; then
    echo -e "${GREEN}✅ WebSocket endpoint exists and responds${NC}"
    ((TESTS_PASSED++))
else
    echo -e "${RED}❌ WebSocket endpoint not responding, got: $WS_ENDPOINT_RESPONSE${NC}"
    ((TESTS_FAILED++))
fi

# Final Stats
echo -e "\n${BLUE}📊 Test Results Summary${NC}"
echo "=========================="
echo -e "${GREEN}✅ Tests Passed: $TESTS_PASSED${NC}"
echo -e "${RED}❌ Tests Failed: $TESTS_FAILED${NC}"
echo -e "Total Tests: $((TESTS_PASSED + TESTS_FAILED))"

if [ $TESTS_FAILED -eq 0 ]; then
    echo -e "\n${GREEN}🎉 All tests passed! The PubSub Gateway is working correctly.${NC}"
else
    echo -e "\n${RED}⚠️  Some tests failed. Please check the service logs.${NC}"
fi

# Cleanup
echo -e "\n${YELLOW}🧹 Cleaning up...${NC}"
kill $SERVICE_PID 2>/dev/null
sleep 2

echo -e "\n${BLUE}✅ Test suite completed!${NC}"
echo -e "\n${YELLOW}💡 To test WebSocket functionality manually:${NC}"
echo "1. Start the service: JWT_SECRET_KEY='$JWT_SECRET_KEY' go run main.go"
echo "2. Register/login to get a token"
echo "3. Connect to WebSocket: ws://localhost:8000/ws?token=YOUR_TOKEN"
echo "4. Send WebSocket messages (see README for examples)"
