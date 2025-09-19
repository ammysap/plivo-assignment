# PubSub System - Low Level Design (LLD)

## Entity Relationship Diagram

```mermaid
erDiagram
    SERVICE {
        string topics "map[string]*Topic"
        string config "Config"
        string startTime "time.Time"
        string mu "sync.RWMutex"
        string shutdown "chan struct{}"
        string wg "sync.WaitGroup"
    }
    
    TOPIC {
        string name "string"
        string subscribers "map[string]*Subscriber"
        string messages "RingBuffer"
        string createdAt "time.Time"
        string mu "sync.RWMutex"
    }
    
    RING_BUFFER {
        string buffer "[]*Message"
        string size "int"
        string head "int"
        string tail "int"
        string count "int"
        string mu "sync.RWMutex"
    }
    
    MESSAGE {
        string id "string"
        string payload "interface{}"
        string topic "string"
        string timestamp "time.Time"
    }
    
    SUBSCRIBER {
        string clientID "string"
        string topicName "string"
        string messageChan "chan *Message"
        string lastSeen "time.Time"
    }
    
    WEBSOCKET_CLIENT {
        string id "string"
        string conn "websocket.Conn"
        string subscriptions "map[string]*Subscriber"
        string mu "sync.RWMutex"
        string done "chan struct{}"
    }
    
    WEBSOCKET_HANDLER {
        string pubsubService "pubsub.Service"
        string clients "map[string]*Client"
        string clientsMu "sync.RWMutex"
        string shutdown "chan struct{}"
    }
    
    CONFIG {
        string ringBufferSize "int"
        string channelBufferSize "int"
    }
    
    USER {
        string id "string"
        string username "string"
        string hashedPassword "string"
        string createdAt "time.Time"
    }
    
    %% Relationships
    SERVICE ||--o{ TOPIC : "manages"
    SERVICE ||--|| CONFIG : "has"
    TOPIC ||--|| RING_BUFFER : "contains"
    TOPIC ||--o{ SUBSCRIBER : "has"
    RING_BUFFER ||--o{ MESSAGE : "stores"
    SUBSCRIBER ||--|| WEBSOCKET_CLIENT : "maps to"
    WEBSOCKET_CLIENT ||--o{ SUBSCRIBER : "subscribes to"
    WEBSOCKET_HANDLER ||--o{ WEBSOCKET_CLIENT : "manages"
    WEBSOCKET_CLIENT ||--|| USER : "authenticated as"
```

## Data Flow Diagram

```mermaid
flowchart TD
    %% Client Layer
    WS[WebSocket Client] --> WH[WebSocket Handler]
    WH --> WC[WebSocket Client]
    
    %% Service Layer
    WC --> PS[PubSub Service]
    PS --> T[Topic]
    T --> RB[Ring Buffer]
    T --> S[Subscriber]
    
    %% Message Flow
    RB --> M[Message]
    M --> MC[Message Channel]
    MC --> WC2[WebSocket Client]
    WC2 --> WS2[WebSocket Connection]
    
    %% User Authentication
    U[User] --> WC
    
    %% Styling
    classDef clientLayer fill:#e1f5fe
    classDef serviceLayer fill:#f3e5f5
    classDef dataLayer fill:#e8f5e8
    classDef messageLayer fill:#fff3e0
    
    class WS,WH,WC,WC2,WS2 clientLayer
    class PS,T,S serviceLayer
    class RB,M,MC dataLayer
    class U messageLayer
```

## System Architecture Overview

```mermaid
graph TB
    subgraph "Client Layer"
        C1[WebSocket Client 1]
        C2[WebSocket Client 2]
        C3[WebSocket Client N]
    end
    
    subgraph "Gateway Layer"
        WH[WebSocket Handler]
        US[User Service]
        TS[Topic Service]
    end
    
    subgraph "Core Layer"
        PS[PubSub Service]
        T1[Topic: orders]
        T2[Topic: payments]
        T3[Topic: notifications]
    end
    
    subgraph "Storage Layer"
        RB1[Ring Buffer 1]
        RB2[Ring Buffer 2]
        RB3[Ring Buffer 3]
        UM[User Memory Store]
    end
    
    %% Connections
    C1 --> WH
    C2 --> WH
    C3 --> WH
    
    WH --> PS
    US --> UM
    TS --> PS
    
    PS --> T1
    PS --> T2
    PS --> T3
    
    T1 --> RB1
    T2 --> RB2
    T3 --> RB3
    
    %% Styling
    classDef client fill:#e3f2fd
    classDef gateway fill:#f1f8e9
    classDef core fill:#fce4ec
    classDef storage fill:#fff8e1
    
    class C1,C2,C3 client
    class WH,US,TS gateway
    class PS,T1,T2,T3 core
    class RB1,RB2,RB3,UM storage
```

## Message Flow Sequence

```mermaid
sequenceDiagram
    participant C as WebSocket Client
    participant WH as WebSocket Handler
    participant PS as PubSub Service
    participant T as Topic
    participant RB as Ring Buffer
    participant S as Subscriber
    participant MC as Message Channel
    
    Note over C,MC: Subscribe Flow
    C->>WH: Subscribe to topic
    WH->>PS: Subscribe(topic, clientID, lastN)
    PS->>T: Add subscriber
    T->>S: Create subscriber
    S->>MC: Create message channel
    PS->>RB: GetLastN(lastN)
    RB-->>PS: Historical messages
    PS->>MC: Send historical messages
    MC-->>C: Deliver historical messages
    
    Note over C,MC: Publish Flow
    C->>WH: Publish message
    WH->>PS: Publish(topic, message)
    PS->>T: Get topic
    T->>RB: Add message
    T->>S: Get all subscribers
    S->>MC: Send to message channel
    MC-->>C: Deliver message
    
    Note over C,MC: Real-time Message Flow
    loop For each subscriber
        MC->>C: Message delivery
    end
```

## Component Interaction Map

```mermaid
graph LR
    subgraph "WebSocket Layer"
        WH[WebSocket Handler]
        WC[WebSocket Client]
    end
    
    subgraph "PubSub Core"
        PS[PubSub Service]
        T[Topic]
        S[Subscriber]
        RB[Ring Buffer]
    end
    
    subgraph "Data Models"
        M[Message]
        U[User]
        C[Config]
    end
    
    %% Primary flows
    WH -.->|manages| WC
    PS -.->|manages| T
    T -.->|contains| RB
    T -.->|has| S
    RB -.->|stores| M
    
    %% Cross-layer connections
    WC -.->|subscribes via| PS
    PS -.->|creates| S
    S -.->|maps to| WC
    
    %% Styling
    classDef websocket fill:#e1f5fe,stroke:#01579b
    classDef pubsub fill:#f3e5f5,stroke:#4a148c
    classDef data fill:#e8f5e8,stroke:#1b5e20
    
    class WH,WC websocket
    class PS,T,S,RB pubsub
    class M,U,C data
```

## Key Relationships Summary

| Relationship | Type | Description |
|--------------|------|-------------|
| SERVICE → TOPIC | 1:N | Service manages multiple topics |
| TOPIC → RING_BUFFER | 1:1 | Each topic has one message history buffer |
| TOPIC → SUBSCRIBER | 1:N | Each topic can have multiple subscribers |
| RING_BUFFER → MESSAGE | 1:N | Ring buffer stores multiple messages |
| SUBSCRIBER → WEBSOCKET_CLIENT | 1:1 | Each subscriber maps to one WebSocket client |
| WEBSOCKET_CLIENT → SUBSCRIBER | 1:N | Each client can subscribe to multiple topics |
| WEBSOCKET_HANDLER → WEBSOCKET_CLIENT | 1:N | Handler manages multiple clients |
| SERVICE → CONFIG | 1:1 | Service has one configuration |
| WEBSOCKET_CLIENT → USER | 1:1 | Each client is authenticated as one user |

## Thread Safety & Concurrency

```mermaid
graph TD
    subgraph "Lock Hierarchy"
        L1[Service.mu - RWMutex]
        L2[Topic.mu - RWMutex]
        L3[RingBuffer.mu - RWMutex]
        L4[WebSocketHandler.clientsMu - RWMutex]
        L5[Client.mu - RWMutex]
        L6[UserService.mu - RWMutex]
    end
    
    subgraph "Concurrent Operations"
        O1[Multiple Readers]
        O2[Single Writer]
        O3[Non-blocking Channels]
        O4[Goroutine Pools]
    end
    
    L1 --> L2
    L2 --> L3
    L4 --> L5
    
    O1 -.->|RLock| L1
    O1 -.->|RLock| L2
    O1 -.->|RLock| L3
    O2 -.->|Lock| L1
    O2 -.->|Lock| L2
    O2 -.->|Lock| L3
    
    classDef lock fill:#ffebee,stroke:#c62828
    classDef operation fill:#e8f5e8,stroke:#2e7d32
    
    class L1,L2,L3,L4,L5,L6 lock
    class O1,O2,O3,O4 operation
```

---
