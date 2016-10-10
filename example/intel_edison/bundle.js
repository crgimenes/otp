define([
    'legacyRegistry',
    './src/EdisonTelemetryServerAdapter',
    './src/EdisonTelemetryInitializer',
    './src/EdisonTelemetryModelProvider'
], function (
    legacyRegistry,
    EdisonTelemetryServerAdapter,
    EdisonTelemetryInitializer,
    EdisonTelemetryModelProvider
) {
    legacyRegistry.register("telemetry", {
    "name": "Edison Telemetry Adapter",
    "extensions": {
        "types": [
            {
                "name": "Intel Edison",
                "key": "Edison.board",
                "cssclass": "icon-object",
            },
            {
                "name": "Subsystem",
                "key": "Edison.subsystem",
                "cssclass": "icon-telemetry-panel",
                "model": { "composition": [] }
            },
            {
                "name": "Measurement",
                "key": "Edison.measurement",
                "cssclass": "icon-telemetry-panel",
                "model": { "telemetry": {} },
                "telemetry": {
                    "source": "Edison.source",
                    "domains": [
                        {
                            "name": "Time",
                            "key": "timestamp"
                        }
                    ]
                }
            }
        ],
        "roots": [
            {
                "id": "Edison:board",
                "priority": "preferred",
                "model": {
                    "type": "Edison.board",
                    "name": "Intel Edison",
                    "composition": []
                }
            }
        ],
        "services": [
            {
                "key": "Edison.adapter",
                "implementation": EdisonTelemetryServerAdapter,
                "depends": [ "$q", "Edison_WS_URL" ]
            }
        ],
        "constants": [
            {
                "key": "Edison_WS_URL",
                "priority": "fallback",
                "value": "ws://localhost:8081"
            }
        ],
        "runs": [
            {
                "implementation": EdisonTelemetryInitializer,
                "depends": [ "Edison.adapter", "objectService" ]
            }
        ],
        "components": [
            {
                "provides": "modelService",
                "type": "provider",
                "implementation": EdisonTelemetryModelProvider,
                "depends": [ "Edison.adapter", "$q" ]
            },
            {
                "provides": "telemetryService",
                "type": "provider",
                "implementation": "EdisonTelemetryProvider.js",
                "depends": [ "Edison.adapter", "$q" ]
            }

        ]
        }
    });
});
