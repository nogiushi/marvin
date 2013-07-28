
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

    $scope.doTransition = function(transition) {
	$.ajax({
	    url: "/post?" + $.now(),
	    type: "POST",
	    cache: false,
	    data: {"do_transition": transition},
	    statusCode: {
		404: function() {
		},
		200: function() {
		}
	    },
	    dataType: "html"
	});
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

}
