app.controller('ModuleCtrl', function($scope, $state, $http) {

    $scope.moduleId = $state.params.id;
    $scope.$on('$ionicView.enter', function() {
        $scope.commands = [];
        $http.get('/modules/' + $state.params.id).then(function(resp) {
            for (var key in resp.data.commands) {
                $scope.commands.push({
                    id: key
                });
            }
        });
    });

    //Execute command
    $scope.onClick = function(cmd) {
        $http.get('/modules/' + $state.params.id + '/' + cmd.id);
    };

});
