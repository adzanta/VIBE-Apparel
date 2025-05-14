document.addEventListener("DOMContentLoaded", function () {
    // Konfirmasi saat menghapus produk
    const deleteButtons = document.querySelectorAll(".btn-delete");
    deleteButtons.forEach(function (btn) {
        btn.addEventListener("click", function (e) {
            if (!confirm("Yakin ingin menghapus produk ini?")) {
                e.preventDefault();
            }
        });
    });
});
