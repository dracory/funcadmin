const { createApp, ref, onMounted } = Vue;

createApp({
    setup() {
        const group = ref({
            id: '',
            slug: '',
            path: '',
            description: '',
            metadata: ''
        });
        const loading = ref(true);
        const saving = ref(false);
        const backUrl = urlGroupManager;

        const loadGroup = async () => {
            loading.value = true;
            try {
                const resp = await fetch(urlLoadGroup, {
                    method: 'GET',
                    headers: { 'Accept': 'application/json' }
                });
                const data = await resp.json();
                if (data.status === 'success') {
                    group.value = {
                        id: data.data.group.id,
                        slug: data.data.group.slug,
                        path: data.data.group.path,
                        description: data.data.group.description || '',
                        metadata: data.data.group.metadata || ''
                    };
                } else {
                    Notiflix.Notify.failure(data.message || 'Failed to load group');
                }
            } catch (err) {
                Notiflix.Notify.failure('Failed to load group');
            } finally {
                loading.value = false;
            }
        };

        const saveGroup = async () => {
            saving.value = true;
            try {
                const resp = await fetch(urlSaveGroup, {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({
                        group_id: group.value.id,
                        description: group.value.description,
                        metadata: group.value.metadata
                    })
                });
                const data = await resp.json();
                if (data.status === 'success') {
                    Notiflix.Notify.success('Group saved successfully');
                } else {
                    Notiflix.Notify.failure(data.message || 'Failed to save group');
                }
            } catch (err) {
                Notiflix.Notify.failure('Failed to save group');
            } finally {
                saving.value = false;
            }
        };

        onMounted(() => {
            loadGroup();
        });

        return {
            group,
            loading,
            saving,
            backUrl,
            saveGroup
        };
    }
}).mount('#groupEditApp');
