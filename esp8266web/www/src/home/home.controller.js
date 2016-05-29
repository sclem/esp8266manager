app.controller('HomeCtrl', function($scope, $state, $http) {

    $scope.modules = [];

    $scope.$on('$ionicView.enter', function() {
        $http.get('/modules').then(function(resp) {
            $scope.modules = resp.data.map(function(item) {
                return {
                    id: item
                };
            });
        });
    });

});
