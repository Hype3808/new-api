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
      <Card className='!rounded-2xl'>
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
    <Card className='!rounded-2xl'>
      {/* Card Header */}
      <div className='flex items-center mb-4'>
        <Avatar size='small' color='orange' className='mr-3 shadow-md'>
          <CalendarCheck size={16} />
        </Avatar>
        <div>
          <Typography.Text className='text-lg font-medium'>
            {t('每日签到')}
          </Typography.Text>
          <div className='text-xs text-gray-600'>
            {t('每日签到获取额度奖励')}
          </div>
        </div>
      </div>

      {/* Check-in Content */}
      <div className='space-y-4'>
        {/* Status Info */}
        <div className='grid grid-cols-2 gap-4'>
          {/* Current Quota */}
          <Card className='!rounded-xl' bodyStyle={{ padding: '12px 16px' }}>
            <div className='flex items-center gap-2 mb-1'>
              <Gift size={14} className='text-orange-500' />
              <Typography.Text size='small' type='tertiary'>
                {t('当前额度')}
              </Typography.Text>
            </div>
            <Typography.Text strong>
              {renderQuota(checkinStatus?.current_quota || userState?.user?.quota || 0)}
            </Typography.Text>
          </Card>

          {/* Reward Quota */}
          <Card className='!rounded-xl' bodyStyle={{ padding: '12px 16px' }}>
            <div className='flex items-center gap-2 mb-1'>
              <Gift size={14} className='text-green-500' />
              <Typography.Text size='small' type='tertiary'>
                {t('签到奖励')}
              </Typography.Text>
            </div>
            <Typography.Text strong className='text-green-600'>
              +{renderQuota(checkinStatus?.reward_quota || 0)}
            </Typography.Text>
          </Card>
        </div>

        {/* Threshold Info */}
        <div className='flex items-center gap-2 text-sm text-gray-500'>
          <AlertCircle size={14} />
          <span>
            {t('额度低于 {{threshold}} 时可签到', {
              threshold: renderQuota(checkinStatus?.quota_threshold || 0),
            })}
          </span>
        </div>

        {/* Last Check-in Time */}
        {checkinStatus?.last_checkin_time > 0 && (
          <div className='flex items-center gap-2 text-sm text-gray-500'>
            <Clock size={14} />
            <span>
              {t('上次签到')}: {formatTime(checkinStatus.last_checkin_time)}
            </span>
          </div>
        )}

        {/* Check-in Button */}
        <div className='pt-2'>
          {hasCheckedIn ? (
            <Tooltip content={getNextCheckinText()}>
              <Button
                type='primary'
                theme='solid'
                disabled
                block
                icon={<IconTick />}
                className='!bg-green-500'
              >
                {t('今日已签到')}
              </Button>
            </Tooltip>
          ) : canCheckin ? (
            <Button
              type='primary'
              theme='solid'
              block
              loading={checkinLoading}
              onClick={handleCheckin}
              icon={<CalendarCheck size={16} />}
              className='!bg-orange-500 hover:!bg-orange-600'
            >
              {t('立即签到')}
            </Button>
          ) : (
            <Tooltip content={checkinStatus?.reason || t('当前额度超过签到阈值')}>
              <Button
                type='primary'
                theme='outline'
                disabled
                block
                icon={<AlertCircle size={16} />}
              >
                {t('暂不可签到')}
              </Button>
            </Tooltip>
          )}
        </div>

        {/* Reason if cannot check-in */}
        {!canCheckin && !hasCheckedIn && checkinStatus?.reason && (
          <div className='text-center text-sm text-orange-500'>
            {checkinStatus.reason}
          </div>
        )}
      </div>
    </Card>
  );
};

export default CheckinCard;
