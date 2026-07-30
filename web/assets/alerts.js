// Alert Message Components
// Reusable function to create DaisyUI alert components

(function() {
    'use strict';

    window.createAlert = function(type, message) {
        const alertTypes = {
            success: {
                bgColor: 'bg-green-50',
                borderColor: 'border-green-500',
                textColor: 'text-green-800',
                iconColor: 'text-green-500',
                icon: 'M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z'
            },
            error: {
                bgColor: 'bg-red-50',
                borderColor: 'border-red-500',
                textColor: 'text-red-800',
                iconColor: 'text-red-500',
                icon: 'M10 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z'
            },
            warning: {
                bgColor: 'bg-yellow-50',
                borderColor: 'border-yellow-500',
                textColor: 'text-yellow-800',
                iconColor: 'text-yellow-500',
                icon: 'M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z'
            },
            info: {
                bgColor: 'bg-blue-50',
                borderColor: 'border-blue-500',
                textColor: 'text-blue-800',
                iconColor: 'text-blue-500',
                icon: 'M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z'
            }
        };

        const alertConfig = alertTypes[type] || alertTypes.info;

        const div = document.createElement('div');
        div.setAttribute('role', 'alert');
        div.className = `flex items-center p-4 mb-4 ${alertConfig.bgColor} border-l-4 ${alertConfig.borderColor} rounded`;

        const svg = document.createElementNS('http://www.w3.org/2000/svg', 'svg');
        svg.setAttribute('class', `flex-shrink-0 w-5 h-5 ${alertConfig.iconColor}`);
        svg.setAttribute('fill', 'none');
        svg.setAttribute('stroke', 'currentColor');
        svg.setAttribute('viewBox', '0 0 24 24');

        const path = document.createElementNS('http://www.w3.org/2000/svg', 'path');
        path.setAttribute('stroke-linecap', 'round');
        path.setAttribute('stroke-linejoin', 'round');
        path.setAttribute('stroke-width', '2');
        path.setAttribute('d', alertConfig.icon);

        svg.appendChild(path);

        const span = document.createElement('span');
        span.className = `ml-3 text-sm font-medium ${alertConfig.textColor}`;
        span.textContent = message;

        div.appendChild(svg);
        div.appendChild(span);

        return div.outerHTML;
    };

    window.showAlert = function(type, message) {
        const messageArea = document.getElementById('message-area');
        if (!messageArea) {
            console.error('message-area not found');
            return;
        }

        messageArea.innerHTML = window.createAlert(type, message);

        // Auto-hide after 5 seconds
        setTimeout(function() {
            messageArea.innerHTML = '';
        }, 5000);
    };

    // Function to check and display alert messages from URL parameters
    function checkAndDisplayAlerts() {
        const urlParams = new URLSearchParams(window.location.search);
        const messageArea = document.getElementById('message-area');

        if (!messageArea) return;

        // Check for success message
        const successMsg = urlParams.get('success');
        if (successMsg) {
            messageArea.innerHTML = window.createAlert('success', successMsg);

            // Clean URL without reloading
            const url = new URL(window.location);
            url.searchParams.delete('success');
            window.history.replaceState({}, '', url);

            // Auto-hide after 5 seconds
            setTimeout(function() {
                messageArea.innerHTML = '';
            }, 5000);
            return;
        }

        // Check for error message
        const errorMsg = urlParams.get('error');
        if (errorMsg) {
            messageArea.innerHTML = window.createAlert('error', errorMsg);

            // Clean URL without reloading
            const url = new URL(window.location);
            url.searchParams.delete('error');
            window.history.replaceState({}, '', url);

            // Auto-hide after 5 seconds
            setTimeout(function() {
                messageArea.innerHTML = '';
            }, 5000);
            return;
        }

        // Check for warning message
        const warningMsg = urlParams.get('warning');
        if (warningMsg) {
            messageArea.innerHTML = window.createAlert('warning', warningMsg);

            // Clean URL without reloading
            const url = new URL(window.location);
            url.searchParams.delete('warning');
            window.history.replaceState({}, '', url);

            // Auto-hide after 5 seconds
            setTimeout(function() {
                messageArea.innerHTML = '';
            }, 5000);
            return;
        }

        // Check for info message
        const infoMsg = urlParams.get('info');
        if (infoMsg) {
            messageArea.innerHTML = window.createAlert('info', infoMsg);

            // Clean URL without reloading
            const url = new URL(window.location);
            url.searchParams.delete('info');
            window.history.replaceState({}, '', url);

            // Auto-hide after 5 seconds
            setTimeout(function() {
                messageArea.innerHTML = '';
            }, 5000);
            return;
        }
    }

    // Intercept htmx redirects to preserve alert messages
    document.addEventListener('htmx:beforeSwap', function(event) {
        const xhr = event.detail.xhr;
        const redirectHeader = xhr.getResponseHeader('HX-Redirect');

        if (redirectHeader) {
            try {
                const url = new URL(redirectHeader, window.location.origin);

                // Check for alert messages in the redirect URL
                const successMsg = url.searchParams.get('success');
                const errorMsg = url.searchParams.get('error');
                const warningMsg = url.searchParams.get('warning');
                const infoMsg = url.searchParams.get('info');

                if (successMsg) {
                    sessionStorage.setItem('alert_type', 'success');
                    sessionStorage.setItem('alert_message', successMsg);
                } else if (errorMsg) {
                    sessionStorage.setItem('alert_type', 'error');
                    sessionStorage.setItem('alert_message', errorMsg);
                } else if (warningMsg) {
                    sessionStorage.setItem('alert_type', 'warning');
                    sessionStorage.setItem('alert_message', warningMsg);
                } else if (infoMsg) {
                    sessionStorage.setItem('alert_type', 'info');
                    sessionStorage.setItem('alert_message', infoMsg);
                }
            } catch (e) {
                console.error('Error parsing redirect URL:', e);
            }
        }
    });

    // Check sessionStorage for pending alert messages
    function checkSessionStorageAlerts() {
        const alertType = sessionStorage.getItem('alert_type');
        const alertMessage = sessionStorage.getItem('alert_message');

        if (alertType && alertMessage) {
            window.showAlert(alertType, alertMessage);

            // Clear the sessionStorage
            sessionStorage.removeItem('alert_type');
            sessionStorage.removeItem('alert_message');
        }
    }

    // Run on DOMContentLoaded
    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', function() {
            checkAndDisplayAlerts();
            checkSessionStorageAlerts();
        });
    } else {
        // DOM is already ready, run immediately
        checkAndDisplayAlerts();
        checkSessionStorageAlerts();
    }
})();
