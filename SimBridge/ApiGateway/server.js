import express from 'express';
import cors from 'cors';
import grpc from '@grpc/grpc-js';
import protoLoader from '@grpc/proto-loader';
import { fileURLToPath } from 'url';
import { dirname, join } from 'path';

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

const PROTO_PATH = join(__dirname, '../Protos/sim_bridge.proto');

const packageDefinition = protoLoader.loadSync(PROTO_PATH, {
    keepCase: true,
    longs: String,
    enums: String,
    defaults: true,
    oneofs: true
});

const simbridge = grpc.loadPackageDefinition(packageDefinition).simbridge;

// Create gRPC client
const client = new simbridge.SimBridgeService('localhost:50051', grpc.credentials.createInsecure());

// Create Express app
const app = express();
app.use(cors());
app.use(express.json());

// Health check
app.get('/api/ping', (req, res) => {
    client.Ping({}, (err, response) => {
        if (err) {
            res.status(500).json({ error: err.message });
        } else {
            res.json(response);
        }
    });
});

// Load study
app.post('/api/load-study', (req, res) => {
    const { studyPath } = req.body;
    client.LoadStudy({ study_path: studyPath }, (err, response) => {
        if (err) {
            res.status(500).json({ error: err.message });
        } else {
            res.json(response);
        }
    });
});

// Get signal values
app.post('/api/get-signals', (req, res) => {
    const { signalNames } = req.body;
    client.GetSignalValues({ signal_names: signalNames }, (err, response) => {
        if (err) {
            res.status(500).json({ error: err.message });
        } else {
            res.json(response);
        }
    });
});

// Run simulation
app.post('/api/run-simulation', (req, res) => {
    const { action, speed } = req.body;
    client.RunSimulation({ action, speed: speed || 1.0 }, (err, response) => {
        if (err) {
            res.status(500).json({ error: err.message });
        } else {
            res.json(response);
        }
    });
});

const PORT = process.env.PORT || 3001;
app.listen(PORT, () => {
    console.log(`SimBridge REST API listening on http://localhost:${PORT}`);
});
