function showForm(type) {
    document.getElementById('form-login').classList.add('hidden');
    document.getElementById('form-register').classList.add('hidden');
    document.getElementById('form-forgot').classList.add('hidden');

    if (type === 'login') {
        document.getElementById('form-login').classList.remove('hidden');
    } else if (type === 'register') {
        document.getElementById('form-register').classList.remove('hidden');
    } else if (type === 'forgot') {
        document.getElementById('form-forgot').classList.remove('hidden');
    }
}
