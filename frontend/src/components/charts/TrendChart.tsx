import { useEffect, useRef } from 'react';
import * as echarts from 'echarts';
import type { Reading } from '../../types/domain';

type Props = { readings: Reading[]; height?: number };

export default function TrendChart({ readings, height = 300 }: Props) {
  const ref = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!ref.current) return;
    const chart = echarts.init(ref.current);
    const grouped = readings.reduce<Record<string, Reading[]>>((all, row) => {
      const key = row.sensor?.type ?? 'unknown';
      (all[key] ??= []).push(row);
      return all;
    }, {});
    chart.setOption({
      tooltip: { trigger: 'axis' },
      legend: { top: 0 },
      grid: { left: 45, right: 18, top: 42, bottom: 45 },
      dataZoom: [{ type: 'inside' }, { type: 'slider' }],
      xAxis: { type: 'time' },
      yAxis: { type: 'value', scale: true },
      series: Object.entries(grouped).map(([name, rows]) => ({
        name,
        type: 'line',
        smooth: true,
        showSymbol: false,
        data: rows
          .slice()
          .sort((a, b) => new Date(a.recordedAt).getTime() - new Date(b.recordedAt).getTime())
          .map((row) => [new Date(row.recordedAt).getTime(), row.value]),
      })),
    });
    const resize = () => chart.resize();
    window.addEventListener('resize', resize);
    return () => {
      window.removeEventListener('resize', resize);
      chart.dispose();
    };
  }, [readings]);

  return <div ref={ref} style={{ height }} aria-label="环境趋势图" />;
}
