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

import React, { useContext, useEffect } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import {
  API,
  showError,
  showSuccess,
  updateAPI,
  setUserData,
} from '../../helpers';
import { UserContext } from '../../context/User';
import Loading from '../common/ui/Loading';

const OAuth2Callback = (props) => {
  const { t } = useTranslation();
  const [searchParams] = useSearchParams();
  const [, userDispatch] = useContext(UserContext);
  const navigate = useNavigate();

  const sendCode = async (code, state) => {
    try {
      const params = { code, state };
      if (props.type === 'discord') {
        params.redirect_uri = `${window.location.origin}/oauth/discord`;
      }
      const { data: resData } = await API.get(`/api/oauth/${props.type}`, {
        params,
      });

      const { success, message, data } = resData;

      if (!success) {
        // OAuth code 只能使用一次，不要重试
        showError(message || t('授权失败'));
        // 检查用户是否已登录（绑定场景）
        const user = localStorage.getItem('user');
        if (user) {
          navigate('/console/personal');
        } else {
          // 未登录用户（登录/注册场景）导航到登录页
          navigate('/login');
        }
        return;
      }

      if (message === 'bind') {
        showSuccess(t('绑定成功！'));
        navigate('/console/personal');
      } else {
        userDispatch({ type: 'login', payload: data });
        localStorage.setItem('user', JSON.stringify(data));
        setUserData(data);
        updateAPI();
        showSuccess(t('登录成功！'));
        navigate('/console/token');
      }
    } catch (error) {
      // 网络错误等，OAuth code 已经被消费，不能重试
      showError(error.message || t('授权失败'));
      // 检查用户是否已登录（绑定场景）
      const user = localStorage.getItem('user');
      if (user) {
        navigate('/console/personal');
      } else {
        // 未登录用户导航到登录页
        navigate('/login');
      }
    }
  };

  useEffect(() => {
    const code = searchParams.get('code');
    const state = searchParams.get('state');

    // 参数缺失直接返回
    if (!code) {
      showError(t('未获取到授权码'));
      navigate('/console/personal');
      return;
    }

    sendCode(code, state);
  }, []);

  return <Loading />;
};

export default OAuth2Callback;
