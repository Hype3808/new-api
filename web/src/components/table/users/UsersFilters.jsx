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

import React, { useRef, useState } from 'react';
import { Form, Button, Collapsible, Typography } from '@douyinfe/semi-ui';
import { IconSearch, IconFilter, IconChevronDown, IconChevronUp } from '@douyinfe/semi-icons';

const UsersFilters = ({
  formInitValues,
  setFormApi,
  searchUsers,
  loadUsers,
  activePage,
  pageSize,
  groupOptions,
  loading,
  searching,
  t,
}) => {
  const formApiRef = useRef(null);
  const [showAdvancedFilters, setShowAdvancedFilters] = useState(false);

  const handleReset = () => {
    if (!formApiRef.current) return;
    formApiRef.current.reset();
    setTimeout(() => {
      loadUsers(1, pageSize);
    }, 100);
  };

  const requestCountModeOptions = [
    { label: t('小于'), value: 'less_than' },
    { label: t('大于'), value: 'more_than' },
  ];

  const sortFieldOptions = [
    { label: t('按 ID'), value: 'id' },
    { label: t('按用户名'), value: 'username' },
    { label: t('按请求次数'), value: 'request_count' },
    { label: t('按总额度'), value: 'total_quota' },
    { label: t('按剩余额度'), value: 'remaining_quota' },
  ];

  const sortOrderOptions = [
    { label: t('升序'), value: 'asc' },
    { label: t('降序'), value: 'desc' },
  ];

  return (
    <Form
      initValues={formInitValues}
      getFormApi={(api) => {
        setFormApi(api);
        formApiRef.current = api;
      }}
      onSubmit={() => {
        searchUsers(1, pageSize);
      }}
      allowEmpty={true}
      autoComplete='off'
      layout='horizontal'
      trigger='change'
      stopValidateWithError={false}
      className='w-full md:w-auto order-1 md:order-2'
    >
      <div className='flex flex-col gap-2 w-full md:w-auto'>
        {/* Main filters row */}
        <div className='flex flex-col md:flex-row items-center gap-2 w-full md:w-auto'>
          <div className='relative w-full md:w-64'>
            <Form.Input
              field='searchKeyword'
              prefix={<IconSearch />}
              placeholder={t('支持搜索用户的 ID、用户名、显示名称和邮箱地址')}
              showClear
              pure
              size='small'
            />
          </div>
          <div className='w-full md:w-48'>
            <Form.Select
              field='searchGroup'
              placeholder={t('选择分组')}
              optionList={groupOptions}
              onChange={(value) => {
                // Group change triggers automatic search
                setTimeout(() => {
                  searchUsers(1, pageSize);
                }, 100);
              }}
              className='w-full'
              showClear
              pure
              size='small'
            />
          </div>
          <Button
            type='tertiary'
            icon={showAdvancedFilters ? <IconChevronUp /> : <IconChevronDown />}
            onClick={() => setShowAdvancedFilters(!showAdvancedFilters)}
            size='small'
            className='w-full md:w-auto'
          >
            {t('高级筛选')}
          </Button>
          <div className='flex gap-2 w-full md:w-auto'>
            <Button
              type='tertiary'
              htmlType='submit'
              loading={loading || searching}
              className='flex-1 md:flex-initial md:w-auto'
              size='small'
            >
              {t('查询')}
            </Button>
            <Button
              type='tertiary'
              onClick={handleReset}
              className='flex-1 md:flex-initial md:w-auto'
              size='small'
            >
              {t('重置')}
            </Button>
          </div>
        </div>

        {/* Advanced filters row */}
        <Collapsible isOpen={showAdvancedFilters}>
          <div className='flex flex-col md:flex-row items-center gap-2 w-full md:w-auto p-2 bg-gray-50 dark:bg-gray-800 rounded-lg'>
            {/* ID Range filters */}
            <div className='flex items-center gap-1 w-full md:w-auto'>
              <Typography.Text size='small' className='whitespace-nowrap'>
                ID {t('范围')}:
              </Typography.Text>
              <Form.InputNumber
                field='idMin'
                placeholder={t('最小')}
                min={1}
                hideButtons
                size='small'
                className='w-20'
                pure
              />
              <Typography.Text size='small'>-</Typography.Text>
              <Form.InputNumber
                field='idMax'
                placeholder={t('最大')}
                min={1}
                hideButtons
                size='small'
                className='w-20'
                pure
              />
            </div>

            {/* Request count filter */}
            <div className='flex items-center gap-1 w-full md:w-auto'>
              <Typography.Text size='small' className='whitespace-nowrap'>
                {t('请求次数')}:
              </Typography.Text>
              <Form.Select
                field='requestCountMode'
                placeholder={t('条件')}
                optionList={requestCountModeOptions}
                className='w-20'
                showClear
                pure
                size='small'
              />
              <Form.InputNumber
                field='requestCount'
                placeholder={t('数值')}
                min={0}
                hideButtons
                size='small'
                className='w-24'
                pure
              />
            </div>

            {/* Sort controls */}
            <div className='flex items-center gap-1 w-full md:w-auto'>
              <Typography.Text size='small' className='whitespace-nowrap'>
                {t('排序')}:
              </Typography.Text>
              <Form.Select
                field='sortBy'
                placeholder={t('字段')}
                optionList={sortFieldOptions}
                onChange={() => {
                  setTimeout(() => {
                    searchUsers(1, pageSize);
                  }, 100);
                }}
                className='w-32'
                pure
                size='small'
              />
              <Form.Select
                field='sortOrder'
                placeholder={t('顺序')}
                optionList={sortOrderOptions}
                onChange={() => {
                  setTimeout(() => {
                    searchUsers(1, pageSize);
                  }, 100);
                }}
                className='w-24'
                pure
                size='small'
              />
            </div>
          </div>
        </Collapsible>
      </div>
    </Form>
  );
};

export default UsersFilters;
