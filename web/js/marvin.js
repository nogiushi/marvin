
angular.module('MarvinApp', ['ui.bootstrap'], function ($interpolateProvider) {
  $interpolateProvider.startSymbol('[[');
  $interpolateProvider.endSymbol(']]');
});

function MarvinCtrl($scope) {
    $scope.state = {};
    $scope.errors = [];
    $scope.connection = null;

    $scope.NewConnection = function() {
        connection = new WebSocket('ws://'+document.location.host+'/state');

        connection.onopen = function () {
        };

        connection.onclose = function (e) {
        };

        connection.onerror = function (error) {
            console.log('WebSocket Error ' + error);
            $scope.$apply(function () {
                $scope.errors.push(error);
            });
        };

        connection.onmessage = function(e) {
            $scope.$apply(function () {
                $scope.state = JSON.parse(e.data);
            });
        };
        $scope.connection = connection;
    };

    $(window).on("pageshow", function() {
        $scope.NewConnection();
    });

    $(window).on("pagehide", function() {
        $scope.connection.close();
    });

    $scope.changeState = function(name, value) {
        var m = {"action": "updateSwitch", "name": name, "value": value};
        $scope.connection.send(JSON.stringify(m));
    };

    $scope.ON = {"on": true};
    $scope.OFF = {"on": false};

    $scope.setHue = function(address, value) {
        var m = {"action": "setHue", "address": address, "value": value};
        $scope.connection.send(JSON.stringify(m));
    };

    $scope.allActivities = function() {
        return Object.keys($scope.state.Activities);
    };

    $scope.allStates = function() {
        if ("States" in $scope.state) {
            return Object.keys($scope.state.States);
        } else {
            return [];
        }
    };

    $scope.updateActivity = function(source, target) {
        $.ajax({
            url: "/activities/?" + $.now(),
            type: "POST",
            cache: false,
            data: {
                "sourceActivity": source,
                "targetActivity": target},
            statusCode: {
                404: function() {
                },
                200: function() {
                }
            },
            dataType: "html"
        });
        $scope.targetActivity = "";
    };

    $scope.nextActivities = function(query, process) {
        return $scope.state.NextActivities;
    };

    $scope.getBrightness = function(state) {
        return Math.round(state.bri / 255 * 100, 0);
    };

    $scope.getStateLabel = function(state) {
        var label = "";
        if (state.on===true) {
            label = label + "On";
        } else if (state.on===false) {
            label = label + "Off";
        }
        if (state.bri) {
            label = label + " " + $scope.getBrightness(state) + "%";
        }
        if (state.alert) {
            label = label + " " + state.alert;
        }
        if (state.transitiontime) {
            label = label + " " + Math.round(state.transitiontime / 10, 1) + "sec";
        }
        return label;
    };

    $scope.getColor = function(state) {
        if (state.colormode===undefined) {
            // xy > ct > hs
            if ("xy" in state) {
                colormode = "xy";
            } else if ("ct" in state) {
                colormode = "ct";
            } else if ("hs" in state) {
                colormode = "hs";
            } else {
                colormode = undefined;
            }
        } else {
            colormode = state.colormode;
        }
        if (colormode=="xy") {
            // Calculate the closest point on the color gamut triangle and use that as xy value See step 6 of color to xy.
            var x = state.xy[0];
            var y = state.xy[1];
            var z = 1.0 - x - y;
            var Y = 0.5; //TODO brightness; // The given brightness value
            var X = (Y / y) * x;
            var Z = (Y / y) * z;

            // Convert to RGB using Wide RGB D65 conversion
            var r = X * 1.612 - Y * 0.203 - Z * 0.302;
            var g = -X * 0.509 + Y * 1.412 + Z * 0.066;
            var b = X * 0.026 - Y * 0.072 + Z * 0.962;

            if (r <= 0.0031308) {
                r = 12.92 * r;
            } else {
                r = (1.0 + 0.055) * Math.pow(r, (1.0 / 2.4)) - 0.055;
            }
            if (g <= 0.0031308) {
                g = 12.92 * g;
            } else {
                g = (1.0 + 0.055) * Math.pow(g, (1.0 / 2.4)) - 0.055;
            }
            if (b <= 0.0031308) {
                b = 12.92 * b;
            } else {
                b = (1.0 + 0.055) * Math.pow(b, (1.0 / 2.4)) - 0.055;
            }
            r = Math.round(r * 255, 0); g = Math.round(g * 255, 0); b = Math.round(b * 255, 0);
            // TODO: getting values bigger than 255
            return "rgb(" + r + "," + g + "," + b + ")";
        } if (colormode=="hs") {
            hue = Math.round(state.hue / 65535 * 360, 2);
            saturation = Math.round(state.sat / 255 * 100, 2);
            brightness = 50; // TODO: Math.round(state["bri"] / 255 * 100, 2);
            return "hsl(" + hue + "," + saturation + "%," + brightness +"%)";
        } if (colormode=="ct") {
            return "white";
        } else {
            return "";
        }
    };

}
