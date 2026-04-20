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

import { useState, useCallback, useEffect } from 'react';
import { useTranslation } from 'react-i18next';
import { API, showError } from '../../helpers';
import { DATE_RANGE_PRESETS } from '../../constants/console.constants';
import dayjs from 'dayjs';

export const useTrendAnalysisData = () => {
  const { t } = useTranslation();

  const [loading, setLoading] = useState(false);
  const [trendData, setTrendData] = useState({
    daily_registrations: [],
    daily_calls: [],
    daily_tokens: [],
  });

  // Default to start of this week to now
  const getDefaultTimeRange = () => {
    const now = dayjs();
    const startOfWeek = now.startOf('week').add(1, 'day'); // Monday as start
    return {
      startTimestamp: Math.floor(startOfWeek.unix()),
      endTimestamp: Math.floor(now.unix()),
    };
  };

  const [formApi, setFormApi] = useState(null);
  const formInitValues = {
    dateRange: [
      dayjs().startOf('week').add(1, 'day').toDate(),
      dayjs().endOf('day').toDate(),
    ],
  };

  const fetchTrendData = useCallback(async (startTimestamp, endTimestamp) => {
    setLoading(true);
    try {
      const res = await API.get('/api/log/trend', {
        params: {
          start_timestamp: startTimestamp,
          end_timestamp: endTimestamp,
        },
      });
      const { success, message, data } = res.data;
      if (success) {
        setTrendData(data || {
          daily_registrations: [],
          daily_calls: [],
          daily_tokens: [],
        });
      } else {
        showError(message || t('获取趋势数据失败'));
      }
    } catch (error) {
      showError(t('获取趋势数据失败') + ': ' + error.message);
    } finally {
      setLoading(false);
    }
  }, [t]);

  const handleRefresh = useCallback(async () => {
    if (!formApi) return;
    const values = formApi.getValues();
    let startTimestamp = 0;
    let endTimestamp = 0;
    if (values.dateRange && values.dateRange.length === 2) {
      startTimestamp = Math.floor(new Date(values.dateRange[0]).getTime() / 1000);
      endTimestamp = Math.floor(new Date(values.dateRange[1]).getTime() / 1000);
    }
    await fetchTrendData(startTimestamp, endTimestamp);
  }, [formApi, fetchTrendData]);

  useEffect(() => {
    // Fetch with default time range on mount
    const defaultRange = getDefaultTimeRange();
    fetchTrendData(defaultRange.startTimestamp, defaultRange.endTimestamp);
  }, []);

  return {
    loading,
    trendData,
    formInitValues,
    setFormApi,
    handleRefresh,
    DATE_RANGE_PRESETS,
    t,
  };
};