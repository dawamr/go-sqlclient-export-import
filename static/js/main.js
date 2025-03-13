// Main JavaScript file for SQL Client

document.addEventListener('DOMContentLoaded', function() {
    // File input handling
    const fileInputs = document.querySelectorAll('input[type="file"]');
    fileInputs.forEach(input => {
        input.addEventListener('change', function(e) {
            const fileName = e.target.files[0]?.name;
            const fileLabel = this.parentElement.querySelector('span');
            if (fileName && fileLabel) {
                fileLabel.textContent = fileName;
            }
        });
    });

    // Form validation
    const forms = document.querySelectorAll('form');
    forms.forEach(form => {
        form.addEventListener('submit', function(e) {
            const requiredFields = form.querySelectorAll('[required]');
            let isValid = true;

            requiredFields.forEach(field => {
                if (!field.value.trim()) {
                    isValid = false;
                    field.classList.add('border-red-500');
                    
                    // Add error message if it doesn't exist
                    let errorMsg = field.parentElement.querySelector('.error-message');
                    if (!errorMsg) {
                        errorMsg = document.createElement('p');
                        errorMsg.className = 'text-red-500 text-xs mt-1 error-message';
                        errorMsg.textContent = 'This field is required';
                        field.parentElement.appendChild(errorMsg);
                    }
                } else {
                    field.classList.remove('border-red-500');
                    const errorMsg = field.parentElement.querySelector('.error-message');
                    if (errorMsg) {
                        errorMsg.remove();
                    }
                }
            });

            if (!isValid) {
                e.preventDefault();
            }
        });
    });

    // Auto-populate port based on database type
    const dbTypeSelects = document.querySelectorAll('select[name="type"]');
    dbTypeSelects.forEach(select => {
        select.addEventListener('change', function() {
            const portInput = this.closest('form').querySelector('input[name="port"]');
            if (portInput) {
                switch (this.value) {
                    case 'mysql':
                    case 'mariadb':
                        portInput.value = '3306';
                        break;
                    case 'postgres':
                        portInput.value = '5432';
                        break;
                }
            }
        });
    });

    // Success message animation
    const successMessages = document.querySelectorAll('.bg-green-100');
    successMessages.forEach(message => {
        message.classList.add('success-message');
        
        // Auto-hide success message after 5 seconds
        setTimeout(() => {
            message.style.opacity = '0';
            message.style.transition = 'opacity 0.5s';
            setTimeout(() => {
                message.style.display = 'none';
            }, 500);
        }, 5000);
    });
}); 