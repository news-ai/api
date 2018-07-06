$(function () {
    // Checking password
    var instance =  $('form').parsley();
    $('form').on('submit',function(){
        if (instance.isValid()) {
            return true;
        }
        return false;
    });
});