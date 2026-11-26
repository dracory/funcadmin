const { createApp, ref, onMounted } = Vue;

createApp({
    setup() {
        const versions = ref([]);
        const loading = ref(true);
        const expandedId = ref(null);

        const loadVersions = async () => {
            loading.value = true;
            try {
                const resp = await fetch(urlVersionsLoad, {
                    method: 'GET',
                    headers: { 'Accept': 'application/json' }
                });
                const data = await resp.json();
                if (data.status === 'success') {
                    versions.value = data.data.versions || [];
                } else {
                    Notiflix.Notify.failure(data.message || 'Failed to load versions');
                }
            } catch (err) {
                Notiflix.Notify.failure('Failed to load versions');
            } finally {
                loading.value = false;
            }
        };

        const toggleView = (id) => {
            expandedId.value = expandedId.value === id ? null : id;
        };

        const getExpandedSource = () => {
            const v = versions.value.find(v => v.id === expandedId.value);
            return v ? v.source : '';
        };

        const truncate = (text, max) => {
            if (!text) return '';
            return text.length > max ? text.substring(0, max) + '...' : text;
        };

        const formatTime = (unixStr) => {
            if (!unixStr) return '';
            const ts = parseInt(unixStr, 10);
            if (isNaN(ts)) return unixStr;
            const d = new Date(ts * 1000);
            return d.toLocaleString();
        };

        const confirmRestore = (v) => {
            Notiflix.Confirm.show(
                'Restore Version',
                'Restore this version? The current source will be saved as a new version first, so you can always come back to it.',
                'Yes, restore it',
                'Cancel',
                async () => {
                    try {
                        const resp = await fetch(urlVersionRestore, {
                            method: 'POST',
                            headers: { 'Content-Type': 'application/json' },
                            body: JSON.stringify({ function_id: versionFunctionId, version_id: v.id })
                        });
                        const data = await resp.json();
                        if (data.status === 'success') {
                            Notiflix.Notify.success('Version restored successfully');
                            loadVersions();
                        } else {
                            Notiflix.Notify.failure(data.message || 'Failed to restore version');
                        }
                    } catch (err) {
                        Notiflix.Notify.failure('Failed to restore version');
                    }
                }
            );
        };

        const confirmDelete = (v) => {
            Notiflix.Confirm.show(
                'Delete Version',
                'Delete this saved version? This cannot be undone.',
                'Yes, delete it',
                'Cancel',
                async () => {
                    try {
                        const resp = await fetch(urlVersionDelete, {
                            method: 'POST',
                            headers: { 'Content-Type': 'application/json' },
                            body: JSON.stringify({ version_id: v.id })
                        });
                        const data = await resp.json();
                        if (data.status === 'success') {
                            Notiflix.Notify.success('Version deleted');
                            loadVersions();
                        } else {
                            Notiflix.Notify.failure(data.message || 'Failed to delete version');
                        }
                    } catch (err) {
                        Notiflix.Notify.failure('Failed to delete version');
                    }
                }
            );
        };

        onMounted(() => {
            loadVersions();
        });

        return {
            versions,
            loading,
            expandedId,
            toggleView,
            getExpandedSource,
            truncate,
            formatTime,
            confirmRestore,
            confirmDelete
        };
    }
}).mount('#versionsApp');
