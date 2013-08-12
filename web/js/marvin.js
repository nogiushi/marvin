angular.module('MarvinApp', ['ui.bootstrap'], function ($interpolateProvider) {
  $interpolateProvider.startSymbol('[[');
  $interpolateProvider.endSymbol(']]');
});

function MarvinCtrl($scope) {
    $scope.state = {};
    $scope.errors = [];
    $scope.connection = null;

    $scope.NewConnection = function() {
        var wsproto = "";
        if (document.location.protocol == "https:") {
            wsproto = "wss";
        } else {
            wsproto = "ws";
        }
        connection = new WebSocket(wsproto+"://"+document.location.host+'/state');

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

    $scope.allMessages = function() {
        var choices = [];
        var states = Object.keys($scope.state.Activities);
        for (var i = 0; i < states.length; i++) {
            choices.push("I am " + states[i]);
        }
        var transition = Object.keys($scope.state.Transitions);
        for (i = 0; i < transition.length; i++) {
            choices.push("do transition " + transition[i]);
        }
        var switches = Object.keys($scope.state.Switch);
        for (i = 0; i < switches.length; i++) {
            if ($scope.state.Switch[switches[i]] === true) {
                choices.push("turn off " + switches[i]);
            } else {
                choices.push("turn on " + switches[i]);
            }
        }
        return choices;
    };

    $scope.allStates = function() {
        if ("States" in $scope.state) {
            return Object.keys($scope.state.States);
        } else {
            return [];
        }
    };

    $scope.sendMessage = function(message) {
        var m = {"message": message};
        $scope.connection.send(JSON.stringify(m));
        $scope.message = "";
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
            // TODO: “bri – This is the brightness of a light from its
            // minimum brightness 0 to its maximum brightness 255
            // (note minimum brightness is not off). This range has
            // been calibrated so there a perceptually similar steps
            // in brightness over the range.
            var bri = 0.5 + (state.bri/255.0) / 4;

            var xyb = {x:state.xy[0], y:state.xy[1], bri: bri};
            xyb = colorConverter.xyBriForModel(xyb, 'LCT001');
            var rgb = colorConverter.xyBriToRgb(xyb);
            return "#"+colorConverter.rgbToHexString(rgb);
        } if (colormode=="hs") {
            hue = Math.round(state.hue / 65535 * 360, 2);
            saturation = Math.round(state.sat / 255 * 100, 2);
            brightness = 50 + 100 * (state.bri/255.0) / 4; // sync with xy
            return "hsl(" + hue + "," + saturation + "%," + brightness +"%)";
        } if (colormode=="ct") {
            return "white";
        } else {
            return "";
        }
    };

    $scope.recentMessages = function(reverse) {
        var rm = $scope.state.RecentMessages;
        var messages = [];
        for (var i=rm.Start; i<rm.End; i++) {
            messages.push(rm.Buffer[i%rm.Buffer.length]);
        }
        if (reverse) {
            messages.reverse();
        }
        return messages;
    };

}
