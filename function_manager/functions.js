const { createApp, ref, onMounted } = Vue;

createApp({
    setup() {
        const functions = ref([]);
        const loading = ref(true);
        const selectAll = ref(false);

        // Create modal state
        const showCreateModal = ref(false);
        const creating = ref(false);
        const createForm = ref({
            name: '',
            alias: '',
            runtime: 'js',
            source: '',
            timeout_ms: 0
        });

        const loadFunctions = async () => {
            loading.value = true;
            try {
                const resp = await fetch(urlFunctionsLoad, {
                    method: 'GET',
                    headers: { 'Accept': 'application/json' }
                });
                const data = await resp.json();
                if (data.status === 'success') {
                    functions.value = (data.data.functions || []).map(fn => ({
                        ...fn,
                        selected: false
                    }));
                } else {
                    Notiflix.Notify.failure(data.message || 'Failed to load functions');
                }
            } catch (err) {
                Notiflix.Notify.failure('Failed to load functions');
            } finally {
                loading.value = false;
            }
        };

        const toggleSelectAll = () => {
            functions.value.forEach(fn => {
                fn.selected = selectAll.value;
            });
        };

        const getUpdateUrl = (id) => {
            return urlFunctionUpdate.replace('FUNCTION_ID_PLACEHOLDER', id);
        };

        const getInvokeUrl = (id) => {
            return urlFunctionInvoke.replace('FUNCTION_ID_PLACEHOLDER', id);
        };

        const confirmDelete = (fn) => {
            Notiflix.Confirm.show(
                'Delete Function',
                'Are you sure you want to delete "' + (fn.name || fn.alias || fn.id) + '"? This cannot be undone.',
                'Yes, delete it',
                'Cancel',
                async () => {
                    try {
                        const resp = await fetch(urlFunctionDelete, {
                            method: 'POST',
                            headers: { 'Content-Type': 'application/json' },
                            body: JSON.stringify({ function_id: fn.id })
                        });
                        const data = await resp.json();
                        if (data.status === 'success') {
                            Notiflix.Notify.success('Function deleted');
                            loadFunctions();
                        } else {
                            Notiflix.Notify.failure(data.message || 'Failed to delete function');
                        }
                    } catch (err) {
                        Notiflix.Notify.failure('Failed to delete function');
                    }
                }
            );
        };

        const openCreateModal = () => {
            createForm.value = { name: '', alias: '', runtime: 'js', source: '', timeout_ms: 0 };
            showCreateModal.value = true;
        };

        const closeCreateModal = () => {
            showCreateModal.value = false;
        };

        const createFunction = async () => {
            if (!createForm.value.source) return;
            creating.value = true;
            try {
                const resp = await fetch(urlFunctionCreate, {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify(createForm.value)
                });
                const data = await resp.json();
                if (data.status === 'success') {
                    Notiflix.Notify.success('Function created successfully');
                    closeCreateModal();
                    loadFunctions();
                } else {
                    Notiflix.Notify.failure(data.message || 'Failed to create function');
                }
            } catch (err) {
                Notiflix.Notify.failure('Failed to create function');
            } finally {
                creating.value = false;
            }
        };

        onMounted(() => {
            loadFunctions();
        });

        return {
            functions,
            loading,
            selectAll,
            toggleSelectAll,
            getUpdateUrl,
            getInvokeUrl,
            confirmDelete,
            showCreateModal,
            creating,
            createForm,
            openCreateModal,
            closeCreateModal,
            createFunction
        };
    }
}).mount('#functionsApp');
