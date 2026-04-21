import React, { useState, useEffect, useMemo } from 'react';
import { Button, Form, Table, Checkbox, Popover, Typography, Space } from '@douyinfe/semi-ui';
import { IconSearch, IconDownload, IconSetting } from '@douyinfe/semi-icons';
import { useTranslation } from 'react-i18next';
import { API, showError } from '../../helpers';
import { DATE_RANGE_PRESETS } from '../../constants/console.constants';

const ALL_DIMENSIONS = [
    { key: 'company_name', label: '公司' },
    { key: 'first_dept_name', label: '一级部门' },
    { key: 'second_dept_name', label: '二级部门' },
    { key: 'third_dept_name', label: '三级部门' },
    { key: 'model_name', label: '模型' },
];

const DEFAULT_DIMENSIONS = ['company_name', 'first_dept_name', 'second_dept_name', 'third_dept_name', 'model_name'];

const DepartmentLogPage = () => {
    const { t } = useTranslation();
    const [logs, setLogs] = useState([]);
    const [loading, setLoading] = useState(false);
    const [total, setTotal] = useState(0);
    const [page, setPage] = useState(1);
    const [pageSize, setPageSize] = useState(10);
    const [sort, setSort] = useState('prompt_tokens DESC');
    const [formApi, setFormApi] = useState(null);
    const [selectedDimensions, setSelectedDimensions] = useState(DEFAULT_DIMENSIONS);

    const dimensionsParam = useMemo(() => selectedDimensions.join(','), [selectedDimensions]);

    const fetchLogs = async (p = page, s = pageSize, st = sort, dims = dimensionsParam) => {
        setLoading(true);
        try {
            let start_timestamp = 0;
            let end_timestamp = 0;
            let company_name = '';
            let first_dept_name = '';
            let second_dept_name = '';
            let third_dept_name = '';

            if (formApi) {
                const values = formApi.getValues();
                if (values.dateRange && values.dateRange.length === 2 && values.dateRange[0]) {
                    start_timestamp = Math.floor(values.dateRange[0].valueOf() / 1000);
                    end_timestamp = Math.floor(values.dateRange[1].valueOf() / 1000);
                }
                company_name = values.company_name || '';
                first_dept_name = values.first_dept_name || '';
                second_dept_name = values.second_dept_name || '';
                third_dept_name = values.third_dept_name || '';
            } else {
                // Default to start of this week to now
                const now = new Date();
                const day = now.getDay() || 7; 
                now.setHours(0, 0, 0, 0);
                const startOfWeek = new Date(now.getTime() - (day - 1) * 24 * 60 * 60 * 1000);
                start_timestamp = Math.floor(startOfWeek.getTime() / 1000);
                end_timestamp = Math.floor(new Date().getTime() / 1000);
            }

            const res = await API.get(
                `/api/log/department?p=${p}&size=${s}&start_timestamp=${start_timestamp}&end_timestamp=${end_timestamp}&company_name=${company_name}&first_dept_name=${first_dept_name}&second_dept_name=${second_dept_name}&third_dept_name=${third_dept_name}&sort=${st}&dimensions=${dims}`
            );
            const { success, message, data } = res.data;
            if (success) {
                setLogs(data.items || []);
                setTotal(data.total || 0);
            } else {
                showError(message);
            }
        } catch (error) {
            showError(error.message);
        } finally {
            setLoading(false);
        }
    };

    const handleTableChange = ({ pagination, filters, sorter }) => {
        if (sorter) {
            const { field, order } = sorter;
            let orderStr = '';
            if (order === 'ascend') {
                orderStr = `${field} ASC`;
            } else if (order === 'descend') {
                orderStr = `${field} DESC`;
            }
            if (orderStr !== sort) {
                setSort(orderStr);
                fetchLogs(1, pageSize, orderStr, dimensionsParam);
                setPage(1);
            }
        }
    };

    const handleExport = () => {
        let start_timestamp = 0;
        let end_timestamp = 0;
        let company_name = '';
        let first_dept_name = '';
        let second_dept_name = '';
        let third_dept_name = '';

        if (formApi) {
            const values = formApi.getValues();
            if (values.dateRange && values.dateRange.length === 2 && values.dateRange[0]) {
                start_timestamp = Math.floor(values.dateRange[0].valueOf() / 1000);
                end_timestamp = Math.floor(values.dateRange[1].valueOf() / 1000);
            }
            company_name = values.company_name || '';
            first_dept_name = values.first_dept_name || '';
            second_dept_name = values.second_dept_name || '';
            third_dept_name = values.third_dept_name || '';
        }

        const url = `/api/log/department/export?start_timestamp=${start_timestamp}&end_timestamp=${end_timestamp}&company_name=${company_name}&first_dept_name=${first_dept_name}&second_dept_name=${second_dept_name}&third_dept_name=${third_dept_name}&dimensions=${dimensionsParam}`;
        
        API.get(url, { responseType: 'blob' }).then(response => {
            const blob = new Blob([response.data], { type: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet' });
            const downloadUrl = window.URL.createObjectURL(blob);
            const a = document.createElement('a');
            a.href = downloadUrl;
            a.download = `department_logs_${new Date().getTime()}.xlsx`;
            document.body.appendChild(a);
            a.click();
            window.URL.revokeObjectURL(downloadUrl);
        }).catch(err => {
            showError('导出失败');
        });
    };

    useEffect(() => {
        // Set default date range to this week
        if (formApi) {
            const now = new Date();
            const day = now.getDay() || 7; 
            const startOfWeek = new Date(now.getFullYear(), now.getMonth(), now.getDate() - day + 1);
            formApi.setValues({ dateRange: [startOfWeek, new Date()] });
        }
        fetchLogs(page, pageSize, sort, dimensionsParam);
    }, [formApi]);

    // Dynamically build columns based on selected dimensions
    const columns = useMemo(() => {
        const dimColumnMap = {
            'company_name': { title: '公司', dataIndex: 'company_name', key: 'company_name' },
            'first_dept_name': { title: '一级部门', dataIndex: 'first_dept_name', key: 'first_dept_name' },
            'second_dept_name': { title: '二级部门', dataIndex: 'second_dept_name', key: 'second_dept_name' },
            'third_dept_name': { title: '三级部门', dataIndex: 'third_dept_name', key: 'third_dept_name' },
            'model_name': { title: '模型名称', dataIndex: 'model_name', key: 'model_name' },
        };

        const cols = [];
        for (const dim of selectedDimensions) {
            if (dimColumnMap[dim]) {
                cols.push(dimColumnMap[dim]);
            }
        }

        cols.push(
            { title: 'Prompt Tokens', dataIndex: 'prompt_tokens', key: 'prompt_tokens', sorter: true },
            { title: 'Complete Tokens', dataIndex: 'complete_tokens', key: 'complete_tokens', sorter: true },
            { title: 'Total Tokens', dataIndex: 'total_tokens', key: 'total_tokens', sorter: true },
            { title: '员工人数', dataIndex: 'employee_count', key: 'employee_count', sorter: true },
            { title: '员工使用人数', dataIndex: 'use_count', key: 'use_count', sorter: true },
            { title: '员工未使用人数', dataIndex: 'not_use_count', key: 'not_use_count', sorter: true },
        );

        return cols;
    }, [selectedDimensions]);

    const handleDimensionChange = (checkedValues) => {
        if (checkedValues.length === 0) {
            showError('请至少选择一个维度');
            return;
        }
        setSelectedDimensions(checkedValues);
    };

    const handleDimensionApply = () => {
        setPage(1);
        const dims = selectedDimensions.join(',');
        fetchLogs(1, pageSize, sort, dims);
    };

    const dimensionPopoverContent = (
        <div style={{ padding: '8px 0', minWidth: 140 }}>
            <div style={{ marginBottom: 8 }}>
                <Typography.Text strong size='small'>选择统计维度</Typography.Text>
            </div>
            <Checkbox.Group
                direction='vertical'
                value={selectedDimensions}
                onChange={handleDimensionChange}
            >
                {ALL_DIMENSIONS.map(dim => (
                    <Checkbox key={dim.key} value={dim.key} style={{ marginBottom: 4 }}>
                        {dim.label}
                    </Checkbox>
                ))}
            </Checkbox.Group>
            <div style={{ marginTop: 8, display: 'flex', justifyContent: 'flex-end' }}>
                <Button size='small' type='primary' onClick={handleDimensionApply}>
                    应用
                </Button>
            </div>
        </div>
    );

    return (
        <div className='mt-[60px] px-2'>
            <div className='mb-4'>
                <Form
                    getFormApi={(api) => setFormApi(api)}
                    onSubmit={() => {
                        setPage(1);
                        fetchLogs(1, pageSize, sort, dimensionsParam);
                    }}
                    layout='vertical'
                >
                    <div className='grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-2'>
                        <div className='col-span-1 lg:col-span-2'>
                            <Form.DatePicker
                                field='dateRange'
                                className='w-full'
                                type='dateTimeRange'
                                placeholder={['开始时间', '结束时间']}
                                showClear
                                pure
                                size='small'
                                presets={DATE_RANGE_PRESETS ? DATE_RANGE_PRESETS.map((preset) => ({
                                    text: t(preset.text),
                                    start: preset.start(),
                                    end: preset.end(),
                                })) : undefined}
                            />
                        </div>
                        <Form.Input field='company_name' prefix={<IconSearch />} placeholder='公司名称' pure size='small' showClear />
                        <Form.Input field='first_dept_name' prefix={<IconSearch />} placeholder='一级部门名称' pure size='small' showClear />
                        <Form.Input field='second_dept_name' prefix={<IconSearch />} placeholder='二级部门名称' pure size='small' showClear />
                        <Form.Input field='third_dept_name' prefix={<IconSearch />} placeholder='三级部门名称' pure size='small' showClear />
                    </div>
                    <div className='flex mt-3 gap-2 justify-end'>
                        <Button type='tertiary' htmlType='submit' loading={loading} size='small'>查询</Button>
                        <Button type='tertiary' onClick={() => {
                            if (formApi) {
                                formApi.reset();
                                setTimeout(() => fetchLogs(1, pageSize, sort, dimensionsParam), 100);
                            }
                        }} size='small'>重置</Button>
                        <Button type='primary' onClick={handleExport} size='small' icon={<IconDownload />}>导出(Excel)</Button>
                        <Popover
                            content={dimensionPopoverContent}
                            trigger='click'
                            position='bottomRight'
                        >
                            <Button type='tertiary' size='small' icon={<IconSetting />}>维度设置</Button>
                        </Popover>
                    </div>
                </Form>
            </div>
            
            <div style={{ background: 'var(--semi-color-bg-1)', borderRadius: '10px', padding: '16px', border: '1px solid var(--semi-color-border)' }}>
                 <Table
                    columns={columns}
                    dataSource={logs}
                    loading={loading}
                    onChange={handleTableChange}
                    pagination={{
                        currentPage: page,
                        pageSize: pageSize,
                        total: total,
                        showSizeChanger: true,
                        pageSizeOptions: [10, 20, 50, 100],
                        onPageChange: c => {
                            setPage(c);
                            fetchLogs(c, pageSize, sort, dimensionsParam);
                        },
                        onPageSizeChange: s => {
                            setPageSize(s);
                            setPage(1);
                            fetchLogs(1, s, sort, dimensionsParam);
                        }
                    }}
                 />
            </div>
        </div>
    );
};

export default DepartmentLogPage;
