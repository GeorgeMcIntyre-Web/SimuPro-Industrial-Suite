using System;
using System.Threading.Tasks;
using Grpc.Core;
using Simbridge;

namespace SimBridge.Server
{
    class Program
    {
        const int Port = 50051;

        static void Main(string[] args)
        {
            var server = new Grpc.Core.Server
            {
                Services = { SimBridgeService.BindService(new SimBridgeImpl()) },
                Ports = { new ServerPort("localhost", Port, ServerCredentials.Insecure) }
            };
            server.Start();

            Console.WriteLine("SimBridge Server listening on port " + Port);
            Console.WriteLine("Press any key to stop the server...");
            Console.ReadKey();

            server.ShutdownAsync().Wait();
        }
    }

    class SimBridgeImpl : SimBridgeService.SimBridgeServiceBase
    {
        // Fallback to mock if Tecnomatix DLLs fail to load or we are not in eMServer
        private bool _useMock = false;

        public SimBridgeImpl()
        {
#if MOCK
            Console.WriteLine("Running in MOCK mode (DLLs not found at build time).");
            _useMock = true;
#else
            try
            {
                // Attempt to access a Tecnomatix type to trigger assembly load
                var type = typeof(Tecnomatix.Engineering.TxApplication);
                Console.WriteLine("Tecnomatix API detected.");
            }
            catch (Exception ex)
            {
                Console.WriteLine($"Tecnomatix API not available: {ex.Message}. Switching to Mock Mode.");
                _useMock = true;
            }
#endif
        }

        public override Task<PingResponse> Ping(Empty request, ServerCallContext context)
        {
            return Task.FromResult(new PingResponse
            {
                Message = _useMock ? "Pong from Mock Server" : "Pong from Tecnomatix Bridge",
                ServerTime = DateTime.Now.ToString()
            });
        }

        public override Task<StudyResponse> LoadStudy(StudyRequest request, ServerCallContext context)
        {
            Console.WriteLine($"Received LoadStudy: {request.StudyPath}");
            
            if (_useMock)
            {
                return Task.FromResult(new StudyResponse
                {
                    Success = true,
                    Message = $"[MOCK] Successfully loaded study: {request.StudyPath}",
                    LoadedEntitiesCount = 100
                });
            }

            try
            {
                // Real Tecnomatix Logic
                // Note: In a real plugin, this must run on the main thread or via Dispatcher
                // For this PoC, we assume we are running standalone or handling threading elsewhere
                // Tecnomatix.Engineering.TxApplication.LoadData(request.StudyPath); 
                // However, TxApplication usually requires running INSIDE Process Simulate.
                // If running as external EXE, we might need to use COM Interop or just fail gracefully.
                
                return Task.FromResult(new StudyResponse
                {
                    Success = true,
                    Message = "Tecnomatix LoadData called (simulated for external exe)",
                    LoadedEntitiesCount = 0
                });
            }
            catch (Exception ex)
            {
                 return Task.FromResult(new StudyResponse
                {
                    Success = false,
                    Message = $"Error loading study: {ex.Message}",
                    LoadedEntitiesCount = 0
                });
            }
        }

        public override Task<SimulationResponse> RunSimulation(SimulationRequest request, ServerCallContext context)
        {
            Console.WriteLine($"Received RunSimulation: {request.Action}");
            return Task.FromResult(new SimulationResponse
            {
                Success = true,
                State = request.Action.ToString()
            });
        }

        public override Task<SignalValues> GetSignalValues(SignalList request, ServerCallContext context)
        {
             var response = new SignalValues();
             foreach(var name in request.SignalNames)
             {
                 response.Values[name] = _useMock ? "1.0" : "0.0"; // Placeholder
             }
             return Task.FromResult(response);
        }

        public override Task<Empty> SetSignalValues(SignalValues request, ServerCallContext context)
        {
            Console.WriteLine($"Setting {request.Values.Count} signals");
            return Task.FromResult(new Empty());
        }
    }
}
