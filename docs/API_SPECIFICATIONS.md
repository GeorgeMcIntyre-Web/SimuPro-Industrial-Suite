# API Specifications - All 88 Repositories

## API Architecture Overview

Total Endpoints: **1,247** across all repositories
- RESTful APIs: 892 endpoints
- GraphQL APIs: 125 queries/mutations
- WebSocket APIs: 87 real-time channels
- gRPC Services: 143 procedures

## Core Repository APIs

### 1. Process-Simulation-Core API

**Base URL**: `/api/v1/process-simulation`
**Authentication**: Bearer Token (JWT)
**Rate Limit**: 1000 req/min

#### Endpoints

```yaml
POST /simulations
  Description: Create new process simulation
  Request:
    {
      "name": "string",
      "type": "discrete|continuous|hybrid",
      "parameters": {
        "timeHorizon": "number (seconds)",
        "timeStep": "number (milliseconds)",
        "convergenceTolerance": "number",
        "maxIterations": "integer"
      },
      "model": {
        "components": "array",
        "connections": "array",
        "constraints": "array"
      }
    }
  Response:
    {
      "simulationId": "uuid",
      "status": "queued|running|completed|failed",
      "estimatedTime": "seconds",
      "queuePosition": "integer"
    }
  Status Codes:
    201: Created
    400: Invalid parameters
    402: License limit exceeded
    503: Service unavailable

GET /simulations/{id}
  Description: Get simulation status and results
  Response:
    {
      "simulationId": "uuid",
      "status": "string",
      "progress": 0-100,
      "results": {
        "throughput": "number",
        "utilization": "object",
        "bottlenecks": "array",
        "costs": "object"
      },
      "logs": "array",
      "warnings": "array",
      "errors": "array"
    }

PUT /simulations/{id}/parameters
  Description: Update running simulation parameters
  Real-time: Yes
  Throttle: 10 updates/second

DELETE /simulations/{id}
  Description: Cancel running simulation

WebSocket /simulations/{id}/stream
  Description: Real-time simulation updates
  Events:
    - progress: {"progress": 0-100, "eta": "seconds"}
    - metrics: {"timestamp": "iso8601", "metrics": {}}
    - complete: {"results": {}}
    - error: {"error": "string", "details": {}}
```

### 2. PLC-Simulation API

**Base URL**: `/api/v1/plc`
**Protocol**: REST + gRPC for real-time

#### REST Endpoints

```typescript
interface PLCSimulationAPI {
  // Program Management
  POST /programs
    body: {
      name: string
      plcType: "S7-300" | "S7-1200" | "S7-1500" | "ControlLogix" | "Modicon"
      program: string (base64 encoded)
      language: "LAD" | "FBD" | "STL" | "SCL" | "ST"
    }
    returns: { programId: string }

  POST /programs/{id}/validate
    returns: {
      valid: boolean
      errors: Array<{
        line: number
        column: number
        severity: "error" | "warning" | "info"
        message: string
      }>
      metrics: {
        cycleTime: number (ms)
        memoryUsage: number (bytes)
        ioCount: number
      }
    }

  POST /programs/{id}/simulate
    body: {
      duration: number (seconds)
      inputs: Record<string, any>
      scanCycle: number (ms)
      hardware: {
        cpu: string
        memory: number
        modules: Array<string>
      }
    }
    returns: { simulationId: string }

  GET /programs/{id}/tags
    returns: Array<{
      name: string
      dataType: string
      address: string
      value: any
      quality: "good" | "bad" | "uncertain"
    }>

  // Real-time Monitoring
  WebSocket /programs/{id}/monitor
    subscribe: {
      tags: string[]
      interval: number (ms)
    }
    events: {
      tagUpdate: { tag: string, value: any, timestamp: number }
      alarm: { type: string, message: string }
      diagnostic: { code: string, details: object }
    }
}
```

#### gRPC Services

```protobuf
service PLCSimulation {
  rpc StreamTagValues(TagSubscription) returns (stream TagValue);
  rpc ExecuteProgram(PLCProgram) returns (ExecutionResult);
  rpc DebugStep(StepRequest) returns (StepResponse);
  rpc GetMemoryDump(MemoryRequest) returns (MemoryDump);
}

message TagValue {
  string name = 1;
  google.protobuf.Any value = 2;
  int64 timestamp = 3;
  string quality = 4;
}
```

### 3. Manufacturing-Simulation API

**Base URL**: `/api/v1/manufacturing`

```yaml
POST /production-lines
  Description: Model production line
  Complexity: High
  Request:
    {
      "layout": {
        "stations": [
          {
            "id": "string",
            "type": "assembly|test|packaging",
            "position": {"x": 0, "y": 0, "z": 0},
            "capacity": "integer",
            "processingTime": "object"
          }
        ],
        "conveyors": [],
        "buffers": []
      },
      "products": [],
      "resources": {
        "operators": "integer",
        "tools": "array",
        "materials": "array"
      },
      "schedule": {
        "shifts": [],
        "maintenance": []
      }
    }

GET /production-lines/{id}/kpis
  Returns:
    {
      "oee": 0-100,
      "throughput": "number",
      "wip": "integer",
      "cycleTime": "seconds",
      "taktTime": "seconds",
      "utilization": {
        "machines": "percent",
        "operators": "percent",
        "materials": "percent"
      },
      "quality": {
        "firstPassYield": "percent",
        "defectRate": "ppm",
        "rework": "percent"
      }
    }

POST /production-lines/{id}/optimize
  Algorithm: Genetic Algorithm | Simulated Annealing | AI
  Returns:
    {
      "improvements": [
        {
          "type": "layout|schedule|resource",
          "change": "object",
          "impact": {
            "throughput": "+15%",
            "cost": "-10%"
          }
        }
      ]
    }
```

### 4. Robotics-Simulation API

```python
# Python Client Example
class RoboticsAPI:
    
    def create_robot(self, robot_config):
        """
        POST /api/v1/robotics/robots
        """
        return self.post('/robots', {
            'brand': 'ABB|KUKA|Fanuc|UR',
            'model': 'IRB6700|KR240|M20iA',
            'payload': 150,  # kg
            'reach': 2.65,   # meters
            'axes': 6,
            'mounting': 'floor|ceiling|wall|angle'
        })
    
    def create_workcell(self, workcell_config):
        """
        POST /api/v1/robotics/workcells
        """
        return self.post('/workcells', {
            'robots': ['robot_id_1', 'robot_id_2'],
            'fixtures': [...],
            'sensors': [...],
            'safety': {
                'fences': [...],
                'light_curtains': [...],
                'emergency_stops': [...]
            }
        })
    
    def plan_path(self, robot_id, waypoints):
        """
        POST /api/v1/robotics/robots/{id}/path-planning
        """
        return self.post(f'/robots/{robot_id}/path-planning', {
            'start': {'joint_values': [0, 0, 0, 0, 0, 0]},
            'waypoints': waypoints,
            'end': {'joint_values': [90, -45, 30, 0, 60, 0]},
            'constraints': {
                'max_velocity': 2.0,  # m/s
                'max_acceleration': 5.0,  # m/s²
                'collision_check': True,
                'singularity_avoidance': True
            },
            'optimizer': 'time|energy|smooth'
        })
    
    def simulate_program(self, robot_id, program):
        """
        POST /api/v1/robotics/robots/{id}/simulate
        """
        response = self.post(f'/robots/{robot_id}/simulate', {
            'program': program,  # KRL, RAPID, or TP
            'speed': 100,  # percentage
            'visualization': True,
            'export_video': True
        })
        return {
            'cycle_time': response['cycle_time'],
            'path_length': response['path_length'],
            'energy_consumption': response['energy'],
            'collisions': response['collisions'],
            'video_url': response['video_url']
        }
```

### 5. Data & Analytics APIs

#### Time-Series-Analytics API

```javascript
// GraphQL Schema
type Query {
  metrics(
    sources: [String!]!
    metrics: [String!]!
    timeRange: TimeRange!
    aggregation: Aggregation
    filters: [Filter!]
  ): MetricResult!
  
  anomalies(
    datasetId: ID!
    sensitivity: Float
    algorithm: AnomalyAlgorithm
  ): [Anomaly!]!
  
  forecast(
    datasetId: ID!
    horizon: Int!
    confidence: Float
  ): ForecastResult!
}

type Mutation {
  createDataset(input: DatasetInput!): Dataset!
  trainModel(datasetId: ID!, config: ModelConfig!): Model!
}

type Subscription {
  metricUpdates(sources: [String!]!): MetricUpdate!
  alertStream(severity: AlertSeverity): Alert!
}

// REST Endpoints
GET /api/v1/analytics/timeseries/{id}/aggregate
  Query Parameters:
    - period: "1m|5m|1h|1d|1w|1M"
    - function: "avg|sum|min|max|count|stddev"
    - fill: "none|null|zero|previous|linear"
  
POST /api/v1/analytics/correlation
  Body:
    {
      "series1": "metric_id",
      "series2": "metric_id",
      "method": "pearson|spearman|kendall",
      "lag": 0-100
    }
```

### 6. Integration Connector APIs

#### OPC-UA-Connector API

```yaml
POST /api/v1/opcua/servers
  Description: Connect to OPC UA server
  Request:
    endpoint: "opc.tcp://192.168.1.100:4840"
    authentication:
      type: "anonymous|username|certificate"
      credentials: {}
    options:
      securityMode: "None|Sign|SignAndEncrypt"
      securityPolicy: "None|Basic128Rsa15|Basic256"
  
POST /api/v1/opcua/servers/{id}/browse
  Description: Browse OPC UA address space
  Returns: Node tree structure

POST /api/v1/opcua/subscriptions
  Description: Create subscription to OPC UA variables
  Request:
    serverId: "string"
    nodes: ["ns=2;i=1234", "ns=2;s=Demo.Tag1"]
    publishingInterval: 1000  # ms
    callback: "webhook_url"

WebSocket /api/v1/opcua/stream
  Description: Real-time OPC UA data stream
  Protocol: ws://
  Authentication: Token in query string
  Format: MessagePack for efficiency
```

#### S7-Connector API

```python
# S7 Protocol Specific Endpoints
class S7ConnectorAPI:
    
    base_url = "/api/v1/s7"
    
    endpoints = {
        "POST /connect": {
            "description": "Connect to S7 PLC",
            "body": {
                "ip": "192.168.0.1",
                "rack": 0,
                "slot": 1,
                "plc_type": "S7-300|S7-400|S7-1200|S7-1500"
            }
        },
        
        "GET /plcs/{id}/info": {
            "description": "Get PLC information",
            "response": {
                "model": "CPU 315-2 PN/DP",
                "serial": "S C-X4U12345",
                "version": "V3.2.12",
                "status": "RUN|STOP|ERROR"
            }
        },
        
        "POST /plcs/{id}/read": {
            "description": "Read data from PLC",
            "body": {
                "area": "DB|MB|IB|QB",
                "db_number": 100,  # if area=DB
                "start": 0,
                "amount": 10,
                "word_len": "Bit|Byte|Word|DWord|Real"
            }
        },
        
        "POST /plcs/{id}/write": {
            "description": "Write data to PLC",
            "body": {
                "area": "DB|MB|QB",
                "db_number": 100,
                "start": 0,
                "data": [1, 2, 3, 4]
            }
        },
        
        "GET /plcs/{id}/alarms": {
            "description": "Get active alarms",
            "response": [
                {
                    "id": 1001,
                    "timestamp": "2024-01-15T10:30:00Z",
                    "message": "Motor overload",
                    "severity": "warning|error|critical"
                }
            ]
        }
    }
```

### 7. Visualization APIs

#### 3D-Visualization API

```typescript
interface Visualization3DAPI {
  // Scene Management
  POST /api/v1/3d/scenes
    body: {
      name: string
      type: "factory|workcell|machine|product"
      units: "mm|cm|m|inch|feet"
      lighting: "default|industrial|outdoor"
    }
    returns: { sceneId: string, viewerUrl: string }

  POST /api/v1/3d/scenes/{id}/models
    body: {
      modelUrl: string  // URL to GLTF/OBJ/STEP file
      position: Vector3
      rotation: Euler
      scale: Vector3
      material?: MaterialConfig
    }

  // Animation
  POST /api/v1/3d/scenes/{id}/animations
    body: {
      targets: string[]  // model IDs
      keyframes: Array<{
        time: number
        position?: Vector3
        rotation?: Euler
        scale?: Vector3
      }>
      duration: number
      loop: boolean
    }

  // Real-time Updates
  WebSocket /api/v1/3d/scenes/{id}/realtime
    send: {
      type: "update_transform"
      modelId: string
      transform: Matrix4
    }
    receive: {
      type: "collision" | "selection" | "measurement"
      data: object
    }

  // Export
  POST /api/v1/3d/scenes/{id}/export
    body: {
      format: "GLTF|OBJ|FBX|USD|VIDEO"
      quality: "low|medium|high|ultra"
      fps?: 30 | 60
    }
    returns: { downloadUrl: string }
}
```

### API Gateway Aggregation

```yaml
# Unified Gateway API that combines multiple services
POST /api/v1/gateway/simulate-production
  Description: Orchestrates multiple simulations
  Combines:
    - Process-Simulation-Core
    - Manufacturing-Simulation
    - Robotics-Simulation
    - Energy-Optimization
  Request:
    {
      "factory": "factory_id",
      "scenario": {
        "products": [],
        "volume": 10000,
        "timeframe": "1_month"
      },
      "optimize": ["throughput", "energy", "cost"],
      "constraints": {}
    }
  Response:
    {
      "results": {
        "throughput": {},
        "energy": {},
        "robotics": {},
        "bottlenecks": [],
        "recommendations": []
      },
      "visualizations": {
        "3d_scene": "url",
        "dashboards": "url",
        "reports": ["url1", "url2"]
      }
    }

# Batch Operations
POST /api/v1/gateway/batch
  Description: Execute multiple API calls in single request
  Request:
    {
      "operations": [
        {
          "id": "op1",
          "method": "POST",
          "url": "/plc/validate",
          "body": {}
        },
        {
          "id": "op2",
          "method": "GET",
          "url": "/simulations/{op1.result.id}",
          "depends": ["op1"]
        }
      ]
    }
```

## API Rate Limits & Quotas

```javascript
const rateLimits = {
  "starter": {
    requests_per_minute: 60,
    requests_per_day: 10000,
    concurrent_simulations: 2,
    max_simulation_time: 3600,  // seconds
    storage_gb: 10,
    api_keys: 2
  },
  
  "professional": {
    requests_per_minute: 600,
    requests_per_day: 100000,
    concurrent_simulations: 10,
    max_simulation_time: 86400,
    storage_gb: 100,
    api_keys: 10
  },
  
  "enterprise": {
    requests_per_minute: 6000,
    requests_per_day: "unlimited",
    concurrent_simulations: 100,
    max_simulation_time: "unlimited",
    storage_gb: 1000,
    api_keys: "unlimited"
  }
};
```

## API Versioning Strategy

```yaml
Versioning Rules:
  - URL versioning: /api/v1/, /api/v2/
  - Breaking changes increment major version
  - Deprecation notice: 6 months
  - Sunset period: 12 months
  - Version header: X-API-Version

Current Versions:
  v1: Stable (Current)
  v2: Beta (Q2 2024)
  v0: Deprecated (Sunset: Q3 2024)

Migration Support:
  - Compatibility layer for 12 months
  - Migration guides
  - SDK updates
  - Webhook notifications
```

## Common API Response Formats

```typescript
// Success Response
{
  "success": true,
  "data": {},
  "metadata": {
    "timestamp": "2024-01-15T10:30:00Z",
    "version": "1.0.0",
    "request_id": "uuid"
  }
}

// Error Response
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid input parameters",
    "details": [
      {
        "field": "timeStep",
        "issue": "Must be positive number",
        "location": "body"
      }
    ]
  },
  "metadata": {
    "timestamp": "2024-01-15T10:30:00Z",
    "request_id": "uuid",
    "support_url": "https://support.simupro.io/errors/VALIDATION_ERROR"
  }
}

// Pagination
{
  "data": [],
  "pagination": {
    "page": 1,
    "per_page": 50,
    "total_pages": 10,
    "total_items": 487,
    "has_next": true,
    "has_previous": false
  },
  "links": {
    "self": "/api/v1/resource?page=1",
    "next": "/api/v1/resource?page=2",
    "last": "/api/v1/resource?page=10"
  }
}
```

## API Authentication & Security

```yaml
Authentication Methods:
  1. API Key:
     Header: X-API-Key
     Format: sp_live_xxxxxxxxxxxxx
     
  2. JWT Bearer Token:
     Header: Authorization: Bearer <token>
     Expiry: 24 hours
     Refresh: /api/v1/auth/refresh
     
  3. OAuth 2.0:
     Flow: Authorization Code
     Scopes: [read, write, admin]
     
  4. mTLS (Enterprise):
     Client certificates required
     CA: Internal PKI

Security Headers:
  X-RateLimit-Limit: 1000
  X-RateLimit-Remaining: 999
  X-RateLimit-Reset: 1234567890
  X-Request-ID: uuid
  X-Response-Time: 123ms
```

## SDK Code Generation

All APIs support automatic SDK generation:
- OpenAPI 3.0 specification available
- Swagger UI at /api-docs
- Postman collections at /api/postman
- SDK generation via openapi-generator

Available SDKs:
- TypeScript/JavaScript (npm)
- Python (pip)
- Go (go get)
- Java (Maven)
- C# (.NET)
- Ruby (gem)
- PHP (Composer)