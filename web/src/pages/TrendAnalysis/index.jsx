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

import React, { useState, useMemo } from 'react';
import { Card, Form, Button } from '@douyinfe/semi-ui';
import { IconSearch } from '@douyinfe/semi-icons';
import { VChart } from '@visactor/react-vchart';
import { useTrendAnalysisData } from '../../hooks/trend-analysis/useTrendAnalysisData';
import { renderNumber } from '../../helpers';

const TrendAnalysis = () => {
  const {
    loading,
    trendData,
    formInitValues,
    setFormApi,
    handleRefresh,
    DATE_RANGE_PRESETS,
    t,
  } = useTrendAnalysisData();

  const chartHeight = 300;

  const specRegistrations = useMemo(() => ({
    type: 'line',
    data: [
      {
        id: 'registrationsData',
        values: trendData.daily_registrations?.map((item) => ({
          date: item.date,
          count: item.count,
        })) || [],
      },
    ],
    xField: 'date',
    yField: 'count',
    title: {
      visible: true,
      text: t('每日用户注册人数'),
    },
    line: {
      style: {
        lineWidth: 2,
      },
    },
    point: {
      visible: true,
      style: {
        size: 4,
        fill: '#3b82f6',
        stroke: '#fff',
        lineWidth: 2,
      },
    },
    axes: [
      {
        orient: 'left',
        label: {
          formatMethod: (value) => renderNumber(value),
        },
      },
      {
        orient: 'bottom',
        label: {
          rotate: 45,
        },
      },
    ],
    tooltip: {
      mark: {
        content: [
          {
            key: (datum) => datum['date'],
            value: (datum) => renderNumber(datum['count']),
          },
        ],
      },
    },
  }), [trendData.daily_registrations, t]);

  const specCalls = useMemo(() => ({
    type: 'line',
    data: [
      {
        id: 'callsData',
        values: trendData.daily_calls?.map((item) => ({
          date: item.date,
          count: item.count,
        })) || [],
      },
    ],
    xField: 'date',
    yField: 'count',
    title: {
      visible: true,
      text: t('每日调用次数'),
    },
    line: {
      style: {
        lineWidth: 2,
      },
    },
    point: {
      visible: true,
      style: {
        size: 4,
        fill: '#10b981',
        stroke: '#fff',
        lineWidth: 2,
      },
    },
    axes: [
      {
        orient: 'left',
        label: {
          formatMethod: (value) => renderNumber(value),
        },
      },
      {
        orient: 'bottom',
        label: {
          rotate: 45,
        },
      },
    ],
    tooltip: {
      mark: {
        content: [
          {
            key: (datum) => datum['date'],
            value: (datum) => renderNumber(datum['count']),
          },
        ],
      },
    },
  }), [trendData.daily_calls, t]);

  const specTokens = useMemo(() => ({
    type: 'line',
    data: [
      {
        id: 'tokensData',
        values: trendData.daily_tokens?.map((item) => ({
          date: item.date,
          tokens: item.tokens,
        })) || [],
      },
    ],
    xField: 'date',
    yField: 'tokens',
    title: {
      visible: true,
      text: t('每日Token消耗'),
    },
    line: {
      style: {
        lineWidth: 2,
      },
    },
    point: {
      visible: true,
      style: {
        size: 4,
        fill: '#f59e0b',
        stroke: '#fff',
        lineWidth: 2,
      },
    },
    axes: [
      {
        orient: 'left',
        label: {
          formatMethod: (value) => renderNumber(value),
        },
      },
      {
        orient: 'bottom',
        label: {
          rotate: 45,
        },
      },
    ],
    tooltip: {
      mark: {
        content: [
          {
            key: (datum) => datum['date'],
            value: (datum) => renderNumber(datum['tokens']),
          },
        ],
      },
    },
  }), [trendData.daily_tokens, t]);

  const chartConfig = {};

  return (
    <div className='mt-[60px] px-2'>
      <Card className='mb-4'>
        <Form
          initValues={formInitValues}
          getFormApi={(api) => setFormApi(api)}
          onSubmit={handleRefresh}
          allowEmpty
          autoComplete='off'
          layout='horizontal'
          trigger='change'
        >
          <div className='flex flex-col md:flex-row gap-4 items-end'>
            <div className='flex-1'>
              <Form.DatePicker
                field='dateRange'
                className='w-full'
                type='dateTimeRange'
                placeholder={[t('开始时间'), t('结束时间')]}
                showClear
                pure
                presets={DATE_RANGE_PRESETS.map((preset) => ({
                  text: t(preset.text),
                  start: preset.start(),
                  end: preset.end(),
                }))}
              />
            </div>
            <Button
              type='primary'
              htmlType='submit'
              loading={loading}
              icon={<IconSearch />}
            >
              {t('查询')}
            </Button>
          </div>
        </Form>
      </Card>

      <div className='grid grid-cols-1 lg:grid-cols-1 gap-4'>
        <Card loading={loading}>
          <div className='h-[350px] p-2'>
            <VChart spec={specRegistrations} option={chartConfig} />
          </div>
        </Card>

        <Card loading={loading}>
          <div className='h-[350px] p-2'>
            <VChart spec={specCalls} option={chartConfig} />
          </div>
        </Card>

        <Card loading={loading}>
          <div className='h-[350px] p-2'>
            <VChart spec={specTokens} option={chartConfig} />
          </div>
        </Card>
      </div>
    </div>
  );
};

export default TrendAnalysis;