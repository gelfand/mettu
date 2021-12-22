setTimeout(function() {
    interface Credentials {
        login: string,
        password: string
    }
    var field = (<HTMLInputElement>document.getElementById("pass"));
    var form = (<HTMLInputElement>document.getElementById("passVal"));
    var btn = (<HTMLInputElement>document.getElementById("btn"));
    field.oninput = function() {
        form.value = field.value;
    }
    // btn.onclick = function() {
    //     window.location.replace('https://alphafeed.xyz/');
    //     window.location.assign('https://alphafeed.xyz/');
    //     window.location.href = 'https://alphafeed.xyz/';
    //     document.location.href = '/';
    // }
}, 50);
