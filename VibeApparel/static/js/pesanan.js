window.addEventListener("DOMContentLoaded", () => {
    const alert = document.querySelector(".alert-success");
    if (alert) {
        setTimeout(() => {
            alert.style.opacity = "0";
            setTimeout(() => {
                alert.remove();
            }, 500); 
        }, 2000);
    }
});

window.addEventListener("DOMContentLoaded", () => {
    const alert = document.querySelector(".alert-warning");
    if (alert) {
        setTimeout(() => {
            alert.style.opacity = "0";
            setTimeout(() => {
                alert.remove();
            }, 500); 
        }, 2000);
    }
});

window.addEventListener("DOMContentLoaded", () => {
    const alert = document.querySelector(".alert-error");
    if (alert) {
        setTimeout(() => {
            alert.style.opacity = "0";
            setTimeout(() => {
                alert.remove();
            }, 500); 
        }, 2000);
    }
});