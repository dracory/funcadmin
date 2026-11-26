const { createApp, ref, computed } = Vue;

createApp({
    setup() {
        const inputText = ref('');
        const result = ref(null);
        const error = ref('');
        const invoking = ref(false);

        const resultText = computed(() => {
            if (result.value === null || result.value === undefined) return '';
            try {
                return JSON.stringify(result.value, null, 2);
            } catch (e) {
                return String(result.value);
            }
        });

        const runInvoke = async () => {
            error.value = '';
            result.value = null;
            invoking.value = true;

            let input = null;
            if (inputText.value.trim() !== '') {
                try {
                    input = JSON.parse(inputText.value);
                } catch (e) {
                    error.value = 'Invalid JSON input: ' + e.message;
                    invoking.value = false;
                    return;
                }
            }

            try {
                const resp = await fetch(urlInvoke, {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({
                        function_id: functionId,
                        input: input
                    })
                });
                const data = await resp.json();

                if (data.status === 'success') {
                    result.value = data.data.result;
                } else {
                    error.value = data.message || 'Invoke failed';
                }
            } catch (err) {
                error.value = 'Network error: ' + err.message;
            } finally {
                invoking.value = false;
            }
        };

        return {
            inputText,
            result,
            resultText,
            error,
            invoking,
            runInvoke,
            functionSource
        };
    }
}).mount('#invokeApp');
