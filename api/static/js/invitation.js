var codeParameter = $.urlParam('code');
var emailParameter = $.urlParam('email');

if (!codeParameter) {
    codeParameter = '';
}

if (!emailParameter) {
    emailParameter = '';
}

var baseUrl = 'https://tabulae.newsai.org/api/auth/registration?code=' + codeParameter + '&email=' + emailParameter;

console.log(baseUrl);

$("#get-started").attr("href", baseUrl)