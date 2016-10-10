/*global define*/

define(
    function () {
        "use strict";

        function EdisonTelemetrySeries(data) {
            return {
                getPointCount: function () {
                    return data.length;
                },
                getDomainValue: function (index) {
                    return (data[index] || {}).timestamp;
                },
                getRangeValue: function (index) {
                    return (data[index] || {}).value;
                }
            };
        }

        return EdisonTelemetrySeries;
    }
);
