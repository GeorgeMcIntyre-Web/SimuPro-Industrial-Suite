const grpc = require('@grpc/grpc-js');
const protoLoader = require('@grpc/proto-loader');
const path = require('path');

const PROTO_PATH = path.join(__dirname, '../Protos/sim_bridge.proto');

const packageDefinition = protoLoader.loadSync(PROTO_PATH, {
    keepCase: true,
    longs: String,
    enums: String,
    defaults: true,
    oneofs: true
});

const simbridge = grpc.loadPackageDefinition(packageDefinition).simbridge;

function main() {
    const client = new simbridge.SimBridgeService('localhost:50051', grpc.credentials.createInsecure());

    console.log("Sending Ping...");
    client.Ping({}, (err, response) => {
        if (err) {
            console.error("Error:", err);
            return;
        }
        console.log("Ping Response:", response);

        console.log("Sending LoadStudy...");
        client.LoadStudy({ study_path: "C:\\Simulations\\Demo.cojt" }, (err, response) => {
            if (err) {
                console.error("Error:", err);
            } else {
                console.log("LoadStudy Response:", response);
            }

            // Test Signal Management
            console.log("Getting Signals...");
            client.GetSignalValues({ signal_names: ["Robot1_Ready", "Line_Start"] }, (err, response) => {
                if (err) console.error(err);
                else console.log("Signals:", response);
            });

            // Test Simulation Control
            console.log("Starting Simulation...");
            client.RunSimulation({ action: "START", speed: 1.0 }, (err, response) => {
                if (err) console.error(err);
                else console.log("Sim Control:", response);
            });
        });
    });
}

main();
