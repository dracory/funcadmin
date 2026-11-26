const { createApp, ref, onMounted } = Vue;

createApp({
    setup() {
        const groups = ref([]);
        const loading = ref(true);
        const selectAll = ref(false);

        // Create modal state
        const showCreateModal = ref(false);
        const creating = ref(false);
        const createForm = ref({
            slug: '',
            parent: '',
            description: '',
            metadata: ''
        });

        const loadGroups = async () => {
            loading.value = true;
            try {
                const resp = await fetch(urlGroupsLoad, {
                    method: 'GET',
                    headers: { 'Accept': 'application/json' }
                });
                const data = await resp.json();
                if (data.status === 'success') {
                    groups.value = (data.data.groups || []).map(g => ({
                        ...g,
                        selected: false
                    }));
                } else {
                    Notiflix.Notify.failure(data.message || 'Failed to load groups');
                }
            } catch (err) {
                Notiflix.Notify.failure('Failed to load groups');
            } finally {
                loading.value = false;
            }
        };

        const toggleSelectAll = () => {
            groups.value.forEach(g => {
                g.selected = selectAll.value;
            });
        };

        const getUpdateUrl = (id) => {
            return urlGroupUpdate.replace('GROUP_ID_PLACEHOLDER', id);
        };

        const confirmDelete = (g) => {
            Notiflix.Confirm.show(
                'Delete Group',
                'Are you sure you want to delete "' + g.slug + '"? This cannot be undone.',
                'Yes, delete it',
                'Cancel',
                async () => {
                    try {
                        const resp = await fetch(urlGroupDelete, {
                            method: 'POST',
                            headers: { 'Content-Type': 'application/json' },
                            body: JSON.stringify({ group_id: g.id })
                        });
                        const data = await resp.json();
                        if (data.status === 'success') {
                            Notiflix.Notify.success('Group deleted');
                            loadGroups();
                        } else {
                            Notiflix.Notify.failure(data.message || 'Failed to delete group');
                        }
                    } catch (err) {
                        Notiflix.Notify.failure('Failed to delete group');
                    }
                }
            );
        };

        const openCreateModal = () => {
            createForm.value = { slug: '', parent: '', description: '', metadata: '' };
            showCreateModal.value = true;
        };

        const closeCreateModal = () => {
            showCreateModal.value = false;
        };

        const createGroup = async () => {
            if (!createForm.value.slug) return;
            creating.value = true;
            try {
                const resp = await fetch(urlGroupCreate, {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify(createForm.value)
                });
                const data = await resp.json();
                if (data.status === 'success') {
                    Notiflix.Notify.success('Group created successfully');
                    closeCreateModal();
                    loadGroups();
                } else {
                    Notiflix.Notify.failure(data.message || 'Failed to create group');
                }
            } catch (err) {
                Notiflix.Notify.failure('Failed to create group');
            } finally {
                creating.value = false;
            }
        };

        onMounted(() => {
            loadGroups();
        });

        return {
            groups,
            loading,
            selectAll,
            toggleSelectAll,
            getUpdateUrl,
            confirmDelete,
            showCreateModal,
            creating,
            createForm,
            openCreateModal,
            closeCreateModal,
            createGroup
        };
    }
}).mount('#groupsApp');
