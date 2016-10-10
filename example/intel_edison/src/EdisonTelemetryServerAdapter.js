/*global define,WebSocket*/

define(
    [],
    function () {
        "use strict";

        function EdisonTelemetryServerAdapter($q, wsUrl) {
            var ws = new WebSocket(wsUrl),
                histories = {},
                listeners = [],
                dictionary = $q.defer();

            // Handle an incoming message from the server
            ws.onmessage = function (event) {
                var message = JSON.parse(event.data);
                console.log("message" + message.type)

                switch (message.type) {
                case "dictionary":
                    console.log("dictionary")
                    dictionary.resolve(message.value);
                    break;
                case "history":
                    console.log("history")
                    histories[message.id].resolve(message);
                    delete histories[message.id];
                    break;
                case "data":
                    console.log("data")
                    listeners.forEach(function (listener) {
                        listener(message);
                    });
                    break;
                }
            };

            // Request dictionary once connection is established
            ws.onopen = function () {
                console.log("Request dictionary once connection is established");

                ws.send("dictionary");
            };

            return {
                dictionary: function () {
                    console.log("dictionary function");

                    return dictionary.promise;
                },
                history: function (id) {
                    console.log("history " + id);
                    histories[id] = histories[id] || $q.defer();
                    ws.send("history " + id);
                    return histories[id].promise;
                },
                subscribe: function (id) {
                    console.log("subscribe " + id);
                    ws.send("subscribe " + id);
                },
                unsubscribe: function (id) {
                    console.log("unsubscribe " + id);
                    ws.send("unsubscribe " + id);
                },
                listen: function (callback) {
                    console.log("listen function");
                    listeners.push(callback);
                }

            };
        }

        return EdisonTelemetryServerAdapter;
    }
);
