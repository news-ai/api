var codeParameter = $.urlParam('code');
console.log(codeParameter);
document.getElementById("code").value = codeParameter;
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