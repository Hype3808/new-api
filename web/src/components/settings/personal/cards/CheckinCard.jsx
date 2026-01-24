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

import React, { useState, useEffect, useCallback } from 'react';
import {
  Card,
  Button,
  Typography,
  Avatar,
  Tooltip,
  Spin,
} from '@douyinfe/semi-ui';
import { IconTick } from '@douyinfe/semi-icons';
import { CalendarCheck, Gift, Clock, AlertCircle } from 'lucide-react';
import { API, showError, showSuccess, renderQuota } from '../../../../helpers';

const CheckinCard = ({ t, userState, onCheckinSuccess }) => {
  const [loading, setLoading] = useState(false);
  const [checkinLoading, setCheckinLoading] = useState(false);
  const [checkinStatus, setCheckinStatus] = useState(null);

  // Fetch check-in status
  const fetchCheckinStatus = useCallback(async () => {
    setLoading(true);
    try {
      const res = await API.get('/api/user/checkin/status');
      const { success, data } = res.data;
      if (success) {
        setCheckinStatus(data);
      } else {
        // If feature is disabled, set a default status
        setCheckinStatus({ enabled: false });
      }
    } catch (error) {
      // Silently fail - feature might not be enabled
      setCheckinStatus({ enabled: false });
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchCheckinStatus();
  }, [fetchCheckinStatus]);

  // Handle check-in action
  const handleCheckin = async () => {
    setCheckinLoading(true);
    try {
      const res = await API.post('/api/user/checkin');
      const { success, message } = res.data;
      if (success) {
        showSuccess(message || t('签到成功'));
        // Refresh status after successful check-in
        await fetchCheckinStatus();
        // Notify parent to refresh user data
        if (onCheckinSuccess) {
          onCheckinSuccess();
        }
      } else {
        showError(message || t('签到失败'));
      }
    } catch (error) {
      showError(t('签到失败，请稍后重试'));
    } finally {
      setCheckinLoading(false);
    }
  };

  // Format timestamp to readable date
  const formatTime = (timestamp) => {
    if (!timestamp) return t('暂无记录');
    const date = new Date(timestamp * 1000);
    return date.toLocaleString();
  };

  // Calculate time until next check-in
  const getNextCheckinText = () => {
    if (!checkinStatus?.next_checkin_time) return '';
    const now = Math.floor(Date.now() / 1000);
    const diff = checkinStatus.next_checkin_time - now;
    if (diff <= 0) return t('现在可以签到');
    
    const hours = Math.floor(diff / 3600);
    const minutes = Math.floor((diff % 3600) / 60);
    
    if (hours > 0) {
      return t('{{hours}}小时{{minutes}}分钟后可签到', { hours, minutes });
    }
    return t('{{minutes}}分钟后可签到', { minutes });
  };

  // Don't render if feature is disabled or still loading
  if (loading) {
    return (
      <Card className='!rounded-2xl text-sm' bodyStyle={{ padding: '16px' }}>
        <div className='flex items-center justify-center py-8'>
          <Spin />
        </div>
      </Card>
    );
  }

  if (!checkinStatus?.enabled) {
    return null;
  }

  const canCheckin = checkinStatus?.can_checkin && !checkinStatus?.has_checked_in;
  const hasCheckedIn = checkinStatus?.has_checked_in;

  return (
    <Card className='!rounded-2xl text-sm' bodyStyle={{ padding: '16px' }}>
      {/* Card Header */}
      <div className='flex items-center justify-between mb-4'>
        <div className='flex items-center'>
          <Avatar size='small' style={{ backgroundColor: '#1890ff' }} className='mr-3'>
            <CalendarCheck size={18} />
          </Avatar>
          <div>
            <Typography.Text className='text-base font-normal leading-tight block'>
              {t('每日签到')}
            </Typography.Text>
            <div className='text-xs text-gray-600 leading-snug'>
              {t('每天可签到一次，领取固定额度奖励')}
            </div>
          </div>
        </div>
      </div>

      {/* Check-in Content */}
      <div className='space-y-2.5'>
        {/* Today's Reward */}
        <div className='flex items-center justify-between'>
          <Typography.Text className='text-sm text-gray-800'>
            {t('签到奖励')}：
            <Typography.Text strong className='ml-1 text-base text-gray-900'>
              {renderQuota(checkinStatus?.reward_quota || 0)}
            </Typography.Text>
          </Typography.Text>
          <Button
            type='tertiary'
            size='small'
            loading={checkinLoading}
            onClick={handleCheckin}
            disabled={!canCheckin || hasCheckedIn}
            className='!rounded-lg !px-6'
          >
            {t('签到')}
          </Button>
        </div>

        {/* Check-in Status */}
        <Typography.Text
          className='block text-[13px] text-gray-600'
          size='small'
          type='secondary'
        >
          {hasCheckedIn ? t('今日已签到') : t('今日未签到')}
        </Typography.Text>

        {/* Threshold Info */}
        <Typography.Text
          className='block text-xs text-gray-500 leading-relaxed'
          size='small'
          type='tertiary'
        >
          {t('当前额度')} {renderQuota(checkinStatus?.current_quota || userState?.user?.quota || 0)}，
          {t('需低于')} {renderQuota(checkinStatus?.quota_threshold || 0)} {t('才可签到')}
        </Typography.Text>
      </div>
    </Card>
  );
};

export default CheckinCard;
