// Modal Response Handler
// Handles htmx responses for modal forms, automatically closing on success
window.handleModalResponse = function(event) {
    const xhr = event.detail.xhr;

    // Se sucesso (2xx status code)
    if (xhr.status >= 200 && xhr.status < 300) {
        // Extrair URL de redirecionamento do response header customizado
        const redirectUrl = xhr.getResponseHeader('X-Redirect-After-Success');

        if (redirectUrl) {
            // Mostrar mensagem de sucesso por 1 segundo
            setTimeout(() => {
                // Fechar modal
                const modal = event.target.closest('dialog');
                if (modal) modal.close();

                // Redirecionar
                window.location.href = redirectUrl;
            }, 1000);
        }
    }
    // Se erro, deixa mensagem visível no modal
};
