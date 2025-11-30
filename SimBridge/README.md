# SimBridge - Legacy to Modern Bridge

## 🎯 Overview
SimBridge is a gRPC-based communication bridge that connects modern React/TypeScript web applications with legacy .NET Framework 4.8 Tecnomatix Process Simulate environments. It enables browser-based applications to interact with industrial simulation software without requiring ActiveX or desktop plugins.

## 🏗️ Architecture

```
┌─────────────────────┐
│  React Frontend     │  Browser (ES6+)
│  :5175 / :5173      │
└──────────┬──────────┘
           │ HTTP REST
           ▼
┌─────────────────────┐
│  API Gateway        │  Node.js + Express
│  :3001              │
└──────────┬──────────┘
           │ gRPC
           ▼
┌─────────────────────┐
│  SimBridge Server   │  .NET 4.8 Console
│  :50051             │
└──────────┬──────────┘
           │ P/Invoke
           ▼
┌─────────────────────┐
│  Tecnomatix API     │  Process Simulate eMServer
│  .NET 4.8 DLLs      │
└─────────────────────┘
```

## 📁 Project Structure

```
SimBridge/
├── ARCHITECTURE.md          # Architecture diagrams
├── Protos/
│   └── sim_bridge.proto    # gRPC service definitions
├── ApiGateway/             # Node.js REST → gRPC bridge
│   ├── server.js
│   └── package.json
├── Server/                  # .NET 4.8 gRPC server
│   ├── Program.cs
│   └── SimBridge.Server.csproj
└── Client/
    └── test-page.html      # Test console UI
```

## 🚀 Quick Start

### Prerequisites
- .NET Framework 4.8 SDK
- Node.js 18+ with npm
- (Optional) Tecnomatix Process Simulate 16.1+

### 1. Build the gRPC Server
```powershell
cd SimBridge/Server
dotnet build
```

### 2. Start the gRPC Server
```powershell
cd SimBridge/Server
dotnet run
```
You should see:
```
Running in MOCK mode (DLLs not found at build time).
SimBridge Server listening on port 50051
Press any key to stop the server...
```

### 3. Start the API Gateway
```powershell
cd SimBridge/ApiGateway
npm install  # First time only
npm start
```
You should see:
```
SimBridge REST API listening on http://localhost:3001
```

### 4. Test the Connection
Open `SimBridge/Client/test-page.html` in your browser, or use curl:
```powershell
curl http://localhost:3001/api/ping
```

## 🔌 API Endpoints

### REST API (API Gateway :3001)

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/ping` | GET | Health check |
| `/api/load-study` | POST | Load a simulation study |
| `/api/get-signals` | POST | Get signal values |
| `/api/run-simulation` | POST | Control simulation (START/STOP/RESET/STEP) |

### Example Usage

#### Ping
```javascript
fetch('http://localhost:3001/api/ping')
  .then(r => r.json())
  .then(data => console.log(data));
```

#### Load Study
```javascript
fetch('http://localhost:3001/api/load-study', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({ studyPath: 'C:\\Studies\\Sample.cojt' })
});
```

#### Get Signals
```javascript
fetch('http://localhost:3001/api/get-signals', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({ 
    signalNames: ['conveyor_speed', 'robot_status', 'cycle_time'] 
  })
});
```

#### Control Simulation
```javascript
fetch('http://localhost:3001/api/run-simulation', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({ action: 'START', speed: 1.0 })
});
```

## 🧪 Mock Mode vs Production Mode

### Mock Mode (Default)
The server automatically switches to Mock Mode when Tecnomatix DLLs are not available. This is perfect for:
- Development on machines without Tecnomatix
- Testing frontend integration
- CI/CD pipelines

### Production Mode
When Tecnomatix DLLs are found at:
```
C:\Program Files\Tecnomatix_16.1\eMPower\
```
The server will attempt to use the real Tecnomatix API.

**Note**: Running gRPC server standalone (outside of Process Simulate) may have threading limitations. For full functionality, consider:
1. Hosting the gRPC server as a Process Simulate plugin (.dll)
2. Using COM Interop to communicate with a running eMServer instance

## 📦 Integration with Frontend Apps

### TypeScript Client (Recommended)
See: `Process-Simulation/AutoFactoryScope/src/frontend/autofactoryscope-web/src/services/SimBridgeClient.ts`

```typescript
import { bridgeClient } from './services/SimBridgeClient';

// Ping
const status = await bridgeClient.ping();

// Load Study
const result = await bridgeClient.loadStudy('C:\\Studies\\Sample.cojt');

// Get Signals
const signals = await bridgeClient.getSignalValues(['conveyor_speed']);

// Control Simulation
await bridgeClient.runSimulation('START', 1.0);
```

### Vanilla JavaScript
```javascript
const API_BASE = 'http://localhost:3001/api';

async function loadStudy(path) {
  const response = await fetch(`${API_BASE}/load-study`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ studyPath: path })
  });
  return response.json();
}
```

## 🛠️ Development

### Adding New gRPC Methods

1. **Update the Proto file** (`Protos/sim_bridge.proto`):
```protobuf
service SimBridgeService {
  rpc GetRobotPosition (RobotRequest) returns (PositionResponse);
}

message RobotRequest {
  string robot_name = 1;
}

message PositionResponse {
  double x = 1;
  double y = 2;
  double z = 3;
}
```

2. **Rebuild the .NET Server** (Grpc.Tools will regenerate code):
```powershell
cd Server
dotnet build
```

3. **Implement the method** in `Program.cs`:
```csharp
public override Task<PositionResponse> GetRobotPosition(
    RobotRequest request, 
    ServerCallContext context)
{
    // Implementation
}
```

4. **Add REST endpoint** to API Gateway (`ApiGateway/server.js`):
```javascript
app.post('/api/get-robot-position', (req, res) => {
    const { robotName } = req.body;
    client.GetRobotPosition({ robot_name: robotName }, (err, response) => {
        if (err) {
            res.status(500).json({ error: err.message });
        } else {
            res.json(response);
        }
    });
});
```

5. **Update TypeScript Client** (`SimBridgeClient.ts`):
```typescript
async getRobotPosition(robotName: string): Promise<PositionResponse> {
    const response = await fetch(`${this.baseUrl}/get-robot-position`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ robotName })
    });
    return response.json();
}
```

## 🔐 Security Considerations

**⚠️ CURRENT STATE: DEVELOPMENT ONLY**

The current implementation uses:
- Insecure gRPC credentials
- CORS enabled for all origins
- No authentication/authorization

**Before Production Deployment:**
1. Implement TLS/SSL for gRPC
2. Add JWT or OAuth2 authentication
3. Restrict CORS to specific origins
4. Add rate limiting
5. Implement input validation and sanitization
6. Add request logging and monitoring

## 🎯 Next Steps

### Phase 1: Core Functionality (Current)
- [x] Set up gRPC server (.NET 4.8)
- [x] Create REST API Gateway (Node.js)
- [x] Build TypeScript client
- [x] Create test page
- [x] Mock mode for development

### Phase 2: Tecnomatix Integration
- [ ] Access Tecnomatix API directly (requires running in eMServer context)
- [ ] Implement signal streaming (bidirectional)
- [ ] Add object manipulation (create/move/delete objects)
- [ ] Support loading multiple studies

### Phase 3: Production Hardening
- [ ] Add authentication (JWT)
- [ ] Enable TLS/SSL
- [ ] Add comprehensive error handling
- [ ] Implement connection pooling
- [ ] Add health monitoring
- [ ] Create deployment scripts

### Phase 4: Advanced Features
- [ ] Real-time signal streaming
- [ ] Simulation event subscriptions
- [ ] 3D visualization data export
- [ ] Performance metrics collection
- [ ] Multi-study management

## 📊 Testing

### Manual Testing
Use the test console: `SimBridge/Client/test-page.html`

### Automated Testing
```powershell
# Test REST endpoints
curl http://localhost:3001/api/ping

# Test with PowerShell
Invoke-RestMethod -Uri "http://localhost:3001/api/ping"
```

### Integration Testing
See: `Process-Simulation/AutoFactoryScope/src/frontend/autofactoryscope-web/src/tests/simbridge.test.ts`

## 🐛 Troubleshooting

### Port Already in Use
```
Error: listen EADDRINUSE: address already in use :::3001
```
**Solution**: Kill existing process or change port in `server.js`

### gRPC Connection Refused
```
Error: 14 UNAVAILABLE: Connection refused
```
**Solution**: Ensure SimBridge Server is running on port 50051

### CORS Error
```
Access to fetch at 'http://localhost:3001' has been blocked by CORS policy
```
**Solution**: API Gateway includes CORS headers, but check browser console for specific errors

## 📚 Resources

- [gRPC Documentation](https://grpc.io/docs/)
- [Protocol Buffers](https://developers.google.com/protocol-buffers)
- [Tecnomatix API Reference](https://docs.sw.siemens.com/documentation/ps/process-simulate/)

## 📄 License

Part of SimuPro Industrial Suite - Internal Use Only

## 👥 Authors

- George McIntyre - Initial architecture and implementation
- AI Assistant (Antigravity) - Development support

---

**Status**: ✅ Proof of Concept Complete | 🚧 Production Hardening Required
