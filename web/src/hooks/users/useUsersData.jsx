/*
Copyright (C) 2025 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact support@quantumnous.com
*/

import { useState, useEffect } from 'react';
import { useTranslation } from 'react-i18next';
import { API, showError, showSuccess } from '../../helpers';
import { ITEMS_PER_PAGE } from '../../constants';
import { useTableCompactMode } from '../common/useTableCompactMode';

export const useUsersData = () => {
  const { t } = useTranslation();
  const [compactMode, setCompactMode] = useTableCompactMode('users');

  // State management
  const [users, setUsers] = useState([]);
  const [loading, setLoading] = useState(true);
  const [activePage, setActivePage] = useState(1);
  const [pageSize, setPageSize] = useState(ITEMS_PER_PAGE);
  const [searching, setSearching] = useState(false);
  const [groupOptions, setGroupOptions] = useState([]);
  const [userCount, setUserCount] = useState(0);

  // Batch operation states
  const [enableBatchOperation, setEnableBatchOperation] = useState(
    localStorage.getItem('enable-batch-operation-users') === 'true',
  );
  const [selectedUsers, setSelectedUsers] = useState([]);

  // Modal states
  const [showAddUser, setShowAddUser] = useState(false);
  const [showEditUser, setShowEditUser] = useState(false);
  const [editingUser, setEditingUser] = useState({
    id: undefined,
  });

  // Form initial values
  const formInitValues = {
    searchKeyword: '',
    searchGroup: '',
    idMin: '',
    idMax: '',
    requestCount: '',
    requestCountMode: '',
    sortBy: 'id',
    sortOrder: 'desc',
  };

  // Form API reference
  const [formApi, setFormApi] = useState(null);

  // Get form values helper function
  const getFormValues = () => {
    const formValues = formApi ? formApi.getValues() : {};
    return {
      searchKeyword: formValues.searchKeyword || '',
      searchGroup: formValues.searchGroup || '',
      idMin: formValues.idMin || '',
      idMax: formValues.idMax || '',
      requestCount: formValues.requestCount || '',
      requestCountMode: formValues.requestCountMode || '',
      sortBy: formValues.sortBy || 'id',
      sortOrder: formValues.sortOrder || 'desc',
    };
  };

  // Set user format with key field
  const setUserFormat = (users) => {
    for (let i = 0; i < users.length; i++) {
      users[i].key = users[i].id;
    }
    setUsers(users);
  };

  // Load users data
  const loadUsers = async (startIdx, pageSize, sortBy = null, sortOrder = null) => {
    setLoading(true);

    // Pull sort settings from form when not explicitly provided
    if (sortBy === null || sortOrder === null) {
      const formValues = getFormValues();
      sortBy = sortBy === null ? formValues.sortBy : sortBy;
      sortOrder = sortOrder === null ? formValues.sortOrder : sortOrder;
    }

    const res = await API.get(
      `/api/user/?p=${startIdx}&page_size=${pageSize}&sort_by=${encodeURIComponent(
        sortBy,
      )}&sort_order=${encodeURIComponent(sortOrder)}`,
    );
    const { success, message, data } = res.data;
    if (success) {
      const newPageData = data.items;
      setActivePage(data.page);
      setUserCount(data.total);
      setUserFormat(newPageData);
    } else {
      showError(message);
    }
    setLoading(false);
  };

  // Search users with keyword, group, and filters
  const searchUsers = async (
    startIdx,
    pageSize,
    searchKeyword = null,
    searchGroup = null,
    idMin = null,
    idMax = null,
    requestCount = null,
    requestCountMode = null,
    sortBy = null,
    sortOrder = null,
  ) => {
    // If no parameters passed, get values from form
    if (searchKeyword === null) {
      const formValues = getFormValues();
      searchKeyword = formValues.searchKeyword;
      searchGroup = formValues.searchGroup;
      idMin = formValues.idMin;
      idMax = formValues.idMax;
      requestCount = formValues.requestCount;
      requestCountMode = formValues.requestCountMode;
      sortBy = formValues.sortBy;
      sortOrder = formValues.sortOrder;
    }

    const hasFilters =
      searchKeyword !== '' ||
      searchGroup !== '' ||
      idMin !== '' ||
      idMax !== '' ||
      (requestCount !== '' && requestCountMode !== '');

    if (!hasFilters) {
      // If no filters, load all users
      await loadUsers(startIdx, pageSize);
      return;
    }

    setSearching(true);
    let url = `/api/user/search?p=${startIdx}&page_size=${pageSize}`;
    if (searchKeyword) url += `&keyword=${encodeURIComponent(searchKeyword)}`;
    if (searchGroup) url += `&group=${encodeURIComponent(searchGroup)}`;
    if (idMin) url += `&id_min=${idMin}`;
    if (idMax) url += `&id_max=${idMax}`;
    if (requestCount && requestCountMode) {
      url += `&request_count=${requestCount}&request_count_mode=${requestCountMode}`;
    }
    if (sortBy) {
      url += `&sort_by=${encodeURIComponent(sortBy)}`;
    }
    if (sortOrder) {
      url += `&sort_order=${encodeURIComponent(sortOrder)}`;
    }

    const res = await API.get(url);
    const { success, message, data } = res.data;
    if (success) {
      const newPageData = data.items;
      setActivePage(data.page);
      setUserCount(data.total);
      setUserFormat(newPageData);
    } else {
      showError(message);
    }
    setSearching(false);
  };

  // Batch enable users
  const batchEnableUsers = async () => {
    if (selectedUsers.length === 0) {
      showError(t('请先选择要启用的用户！'));
      return;
    }
    setLoading(true);
    const ids = selectedUsers.map((user) => user.id);
    const res = await API.post('/api/user/batch', { ids, action: 'enable' });
    const { success, message, data } = res.data;
    if (success) {
      showSuccess(t('已启用 ${count} 个用户！').replace('${count}', data));
      await refresh();
      setSelectedUsers([]);
    } else {
      showError(message);
    }
    setLoading(false);
  };

  // Batch disable users
  const batchDisableUsers = async () => {
    if (selectedUsers.length === 0) {
      showError(t('请先选择要禁用的用户！'));
      return;
    }
    setLoading(true);
    const ids = selectedUsers.map((user) => user.id);
    const res = await API.post('/api/user/batch', { ids, action: 'disable' });
    const { success, message, data } = res.data;
    if (success) {
      showSuccess(t('已禁用 ${count} 个用户！').replace('${count}', data));
      await refresh();
      setSelectedUsers([]);
    } else {
      showError(message);
    }
    setLoading(false);
  };

  // Manage user operations (promote, demote, enable, disable, delete)
  const manageUser = async (userId, action, record) => {
    // Trigger loading state to force table re-render
    setLoading(true);

    const res = await API.post('/api/user/manage', {
      id: userId,
      action,
    });

    const { success, message } = res.data;
    if (success) {
      showSuccess('操作成功完成！');
      const user = res.data.data;

      // Create a new array and new object to ensure React detects changes
      const newUsers = users.map((u) => {
        if (u.id === userId) {
          if (action === 'delete') {
            return { ...u, DeletedAt: new Date() };
          }
          return { ...u, status: user.status, role: user.role };
        }
        return u;
      });

      setUsers(newUsers);
    } else {
      showError(message);
    }

    setLoading(false);
  };

  const resetUserPasskey = async (user) => {
    if (!user) {
      return;
    }
    try {
      const res = await API.delete(`/api/user/${user.id}/reset_passkey`);
      const { success, message } = res.data;
      if (success) {
        showSuccess(t('Passkey 已重置'));
      } else {
        showError(message || t('操作失败，请重试'));
      }
    } catch (error) {
      showError(t('操作失败，请重试'));
    }
  };

  const resetUserTwoFA = async (user) => {
    if (!user) {
      return;
    }
    try {
      const res = await API.delete(`/api/user/${user.id}/2fa`);
      const { success, message } = res.data;
      if (success) {
        showSuccess(t('二步验证已重置'));
      } else {
        showError(message || t('操作失败，请重试'));
      }
    } catch (error) {
      showError(t('操作失败，请重试'));
    }
  };

  // Handle page change
  const handlePageChange = (page) => {
    setActivePage(page);
    const {
      searchKeyword,
      searchGroup,
      idMin,
      idMax,
      requestCount,
      requestCountMode,
      sortBy,
      sortOrder,
    } = getFormValues();
    const hasFilters =
      searchKeyword !== '' ||
      searchGroup !== '' ||
      idMin !== '' ||
      idMax !== '' ||
      (requestCount !== '' && requestCountMode !== '');

    if (!hasFilters) {
      loadUsers(page, pageSize).then();
    } else {
      searchUsers(
        page,
        pageSize,
        searchKeyword,
        searchGroup,
        idMin,
        idMax,
        requestCount,
        requestCountMode,
        sortBy,
        sortOrder,
      ).then();
    }
  };

  // Handle page size change
  const handlePageSizeChange = async (size) => {
    localStorage.setItem('page-size', size + '');
    setPageSize(size);
    setActivePage(1);
    const { sortBy, sortOrder } = getFormValues();
    loadUsers(activePage, size, sortBy, sortOrder)
      .then()
      .catch((reason) => {
        showError(reason);
      });
  };

  // Handle table row styling for disabled/deleted users
  const handleRow = (record, index) => {
    if (record.DeletedAt !== null || record.status !== 1) {
      return {
        style: {
          background: 'var(--semi-color-disabled-border)',
        },
      };
    } else {
      return {};
    }
  };

  // Refresh data
  const refresh = async (page = activePage) => {
    const {
      searchKeyword,
      searchGroup,
      idMin,
      idMax,
      requestCount,
      requestCountMode,
      sortBy,
      sortOrder,
    } = getFormValues();
    const hasFilters =
      searchKeyword !== '' ||
      searchGroup !== '' ||
      idMin !== '' ||
      idMax !== '' ||
      (requestCount !== '' && requestCountMode !== '');

    if (!hasFilters) {
      await loadUsers(page, pageSize, sortBy, sortOrder);
    } else {
      await searchUsers(
        page,
        pageSize,
        searchKeyword,
        searchGroup,
        idMin,
        idMax,
        requestCount,
        requestCountMode,
        sortBy,
        sortOrder,
      );
    }
  };

  // Fetch groups data
  const fetchGroups = async () => {
    try {
      let res = await API.get(`/api/group/`);
      if (res === undefined) {
        return;
      }
      setGroupOptions(
        res.data.data.map((group) => ({
          label: group,
          value: group,
        })),
      );
    } catch (error) {
      showError(error.message);
    }
  };

  // Modal control functions
  const closeAddUser = () => {
    setShowAddUser(false);
  };

  const closeEditUser = () => {
    setShowEditUser(false);
    setEditingUser({
      id: undefined,
    });
  };

  // Initialize data on component mount
  useEffect(() => {
    loadUsers(0, pageSize)
      .then()
      .catch((reason) => {
        showError(reason);
      });
    fetchGroups().then();
  }, []);

  return {
    // Data state
    users,
    loading,
    activePage,
    pageSize,
    userCount,
    searching,
    groupOptions,

    // Batch operation state
    enableBatchOperation,
    setEnableBatchOperation,
    selectedUsers,
    setSelectedUsers,

    // Modal state
    showAddUser,
    showEditUser,
    editingUser,
    setShowAddUser,
    setShowEditUser,
    setEditingUser,

    // Form state
    formInitValues,
    formApi,
    setFormApi,

    // UI state
    compactMode,
    setCompactMode,

    // Actions
    loadUsers,
    searchUsers,
    manageUser,
    resetUserPasskey,
    resetUserTwoFA,
    batchEnableUsers,
    batchDisableUsers,
    handlePageChange,
    handlePageSizeChange,
    handleRow,
    refresh,
    closeAddUser,
    closeEditUser,
    getFormValues,

    // Translation
    t,
  };
};
