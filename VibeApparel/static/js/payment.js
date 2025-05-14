const buktiTransferInput = document.getElementById('buktiTransfer');
const submitButton = document.getElementById('submitButton');

buktiTransferInput.addEventListener('change', function() {
  if (buktiTransferInput.files.length > 0) {
    submitButton.disabled = false;
  } else {
    submitButton.disabled = true;
  }
});
