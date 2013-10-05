var myModule = angular.module('MarvinApp', ['ui.bootstrap'], function ($interpolateProvider) {
  $interpolateProvider.startSymbol('[[');
  $interpolateProvider.endSymbol(']]');
});

// configure existing services inside initialization blocks.
myModule.config(function($compileProvider) {
  // configure new 'compile' directive by passing a directive
  // factory function. The factory function injects the '$compile'
  $compileProvider.directive('compile', function($compile) {

    // directive factory creates a link function
    return function(scope, element, attrs) {
      scope.$watch(
        function(scope) {
           // watch the 'compile' expression for changes
          return scope.$eval(attrs.compile);
        },
        function(value) {
          // when the 'compile' expression changes
          // assign it into the current DOM
          element.html(value);
 
          // compile the new DOM and link it to the current
          // scope.
          // NOTE: we only compile .childNodes so that
          // we don't get into infinite loop compiling ourselves
          $compile(element.contents())(scope);
        }
      );
    };
  });
});


function MarvinCtrl($scope) {
    $scope.state = {};
    $scope.errors = [];
    $scope.connection = null;
    $scope.messageconnection = null;

    $scope.NewMessageConnection = function() {
        var wsproto = "";
        if (document.location.protocol == "https:") {
            wsproto = "wss";
        } else {
            wsproto = "ws";
        }
        messageconnection = new WebSocket(wsproto+"://"+document.location.host+'/message');

        messageconnection.onopen = function () {
            $scope.messageconnection = messageconnection;
        };

        messageconnection.onclose = function (e) {
            $scope.messageconnection = null;
        };

        messageconnection.onerror = function (error) {
            console.log('WebSocket Error ' + error);
            $scope.$apply(function () {
                $scope.errors.push(error);
            });
        };

        messageconnection.onmessage = function(e) {
            $scope.$apply(function () {
                var msg = JSON.parse(e.data);
                if (msg.Why == "statechanged") {
                    if (msg.Who == "Nog") {
                        $scope.state = JSON.parse(msg.What);
                    }
                } else {
                    $scope.displayMessage(msg);
                }
            });
        };
    };

    $(window).on("pageshow", function() {
        $scope.NewMessageConnection();
    });

    $(window).on("pagehide", function() {
        if ($scope.messageconnection !== null) {
            $scope.messageconnection.close();
        }
        if ($scope.connection !== null) {
            $scope.connection.close();
        }
    });

    $scope.changeState = function(name, value, why) {
        if (value === true) {
            $scope.sendMessage("turn on " + name, "switches", why);
        } else {
            $scope.sendMessage("turn off " + name, "switches", why);
        }
    };

    $scope.ON = {"on": true};
    $scope.OFF = {"on": false};

    $scope.setHue = function(address, value, why) {
        $scope.sendMessage("set hue address " + address + " to " + JSON.stringify(value), why);
    };

    $scope.allMessages = function() {
        var choices = [];
        var states = Object.keys($scope.state.Activities);
        for (var i = 0; i < states.length; i++) {
            choices.push("I am " + states[i]);
        }
        var action = Object.keys($scope.state.Actions);
        for (i = 0; i < action.length; i++) {
            choices.push("do " + action[i]);
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

    $scope.displayMessage = function(message) {
        var u = $('<li class="list-group-item message">' + message.Who + ': <span class="what ' + message.Why + '">' + message.What + '</span> </li>');
        u.append(message);
        u.hide();
        $("#messageinputitem").before(u);
        u.slideDown().delay(10000).animate({opacity: 0}, {duration: 500, always: function() { $(this).remove(); }});
        $scope.displayMessageThen(message);
    };

    $scope.displayMessageThen = function(message) {
        var uu = $('<li class="list-group-item message">' + message.Who + ': <span class="what ' + message.Why + '">' + message.What + '</span> </li>');
        $("#messagesthen").prepend(uu);
    };

    $.getJSON("/messages", function(messages) {
        var length = messages.length;
        for (var i = 0; i < length; i++) {
            if (messages[i].Why != "statechanged") {
                $scope.displayMessageThen(messages[i]);
            }
        }
    });

    $scope.sendMessage = function(message, why) {
        var m = {"message": message, "why": why};
        if ($scope.messageconnection !== null) {
            if ($scope.messageconnection.readyState == 1) {
                $scope.messageconnection.send(JSON.stringify(m));
            } else {
                $scope.errors.push("not ready");
            }
            $scope.message = "";
        }
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
        if (rm !== undefined) {
            for (var i=rm.Start; i<rm.End; i++) {
                messages.push(rm.Buffer[i%rm.Buffer.length]);
            }
            if (reverse) {
                messages.reverse();
            }
        }
        return messages;
    };

    $scope.formatWhen = function(when) {
        return when.substring(11,19);
    };

    $scope.nowThenFlip = function() {
        $("#nowthen").toggleClass("flip");
    };
}
