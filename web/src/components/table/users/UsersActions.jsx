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

import React from 'react';
import { Button, Switch, Typography, Modal } from '@douyinfe/semi-ui';

const UsersActions = ({
  setShowAddUser,
  enableBatchOperation,
  setEnableBatchOperation,
  selectedUsers,
  batchEnableUsers,
  batchDisableUsers,
  loading,
  t,
}) => {
  // Add new user
  const handleAddUser = () => {
    setShowAddUser(true);
  };

  // Handle batch enable
  const handleBatchEnable = () => {
    if (selectedUsers.length === 0) {
      return;
    }
    Modal.confirm({
      title: t('确认启用'),
      content: t('确定要启用选中的 ${count} 个用户吗？').replace(
        '${count}',
        selectedUsers.length,
      ),
      onOk: () => batchEnableUsers(),
    });
  };

  // Handle batch disable
  const handleBatchDisable = () => {
    if (selectedUsers.length === 0) {
      return;
    }
    Modal.confirm({
      title: t('确认禁用'),
      content: t('确定要禁用选中的 ${count} 个用户吗？').replace(
        '${count}',
        selectedUsers.length,
      ),
      okType: 'danger',
      onOk: () => batchDisableUsers(),
    });
  };

  return (
    <div className='flex flex-col gap-2 w-full md:w-auto order-2 md:order-1'>
      <div className='flex flex-wrap items-center gap-2'>
        <Button
          className='w-full md:w-auto'
          onClick={handleAddUser}
          size='small'
        >
          {t('添加用户')}
        </Button>

        {enableBatchOperation && (
          <>
            <Button
              size='small'
              disabled={selectedUsers.length === 0}
              onClick={handleBatchEnable}
              loading={loading}
            >
              {t('批量启用')}
            </Button>
            <Button
              size='small'
              type='danger'
              disabled={selectedUsers.length === 0}
              onClick={handleBatchDisable}
              loading={loading}
            >
              {t('批量禁用')}
            </Button>
          </>
        )}

        <div className='flex items-center gap-2 ml-auto md:ml-2'>
          <Typography.Text strong className='text-sm'>
            {t('批量操作')}
          </Typography.Text>
          <Switch
            size='small'
            checked={enableBatchOperation}
            onChange={(v) => {
              localStorage.setItem('enable-batch-operation-users', v + '');
              setEnableBatchOperation(v);
            }}
          />
        </div>
      </div>
    </div>
  );
};

export default UsersActions;
