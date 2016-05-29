app.controller('HomeCtrl', function($scope, $state, $http, $ionicHistory) {

    $scope.modules = [];

    $scope.$on('$ionicView.afterEnter', function() {
        $ionicHistory.clearHistory();
    });

    $scope.$on('$ionicView.loaded', function() {
        $http.get('/modules').then(function(resp) {
            $scope.modules = resp.data.map(function(item) {
                return {
                    id: item
                };
            });
        });
    });

});
