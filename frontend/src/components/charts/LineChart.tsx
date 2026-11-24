import {useMemo} from 'react';
import {Box, Typography, useTheme} from '@mui/material';

export type Pt = {t: number; v: number};

export default function LineChart({title, points, max, format, color, windowMs}: {
    title: string; points: Pt[]; max: number; format: (v: number) => string; color?: string; windowMs: number;
}) {
    const theme = useTheme();
    const stroke = color ?? theme.palette.primary.main;
    const W = 600, H = 120;
    const {path, area, last, peak} = useMemo(() => {
        if (!points.length) return {path: '', area: '', last: 0, peak: 0};
        const end = points[points.length - 1].t;
        const start = end - windowMs;
        const y = (v: number) => H - 4 - (Math.min(v, max) / max) * (H - 8);
        const x = (t: number) => ((t - start) / windowMs) * W;
        const pts = points.map(p => `${x(p.t).toFixed(1)},${y(p.v).toFixed(1)}`);
        const first = x(points[0].t).toFixed(1);
        return {
            path: 'M' + pts.join(' L'),
            area: `M${first},${H} L${pts.join(' L')} L${x(end).toFixed(1)},${H} Z`,
            last: points[points.length - 1].v,
            peak: Math.max(...points.map(p => p.v)),
        };
    }, [points, max, windowMs]);
    return (
        <Box sx={{p: 1.5, border: 1, borderColor: 'divider', borderRadius: 1, bgcolor: 'background.default'}}>
            <Box sx={{display: 'flex', justifyContent: 'space-between', mb: 0.5}}>
                <Typography variant="subtitle2" sx={{fontWeight: 600}}>{title}</Typography>
                <Typography variant="caption" sx={{color: 'text.secondary'}}>now {format(last)} / peak {format(peak)}</Typography>
            </Box>
            <svg viewBox={`0 0 ${W} ${H}`} width="100%" height={H} preserveAspectRatio="none" style={{display: 'block'}}>
                {[0.25, 0.5, 0.75].map(f => <line key={f} x1={0} x2={W} y1={H * f} y2={H * f} stroke={theme.palette.divider} strokeWidth={1} vectorEffect="non-scaling-stroke"/>)}
                {path && <path d={area} fill={stroke} fillOpacity={0.15}/>}
                {path && <path d={path} fill="none" stroke={stroke} strokeWidth={1.5} vectorEffect="non-scaling-stroke"/>}
            </svg>
            <Box sx={{display: 'flex', justifyContent: 'space-between'}}>
                <Typography variant="caption" sx={{color: 'text.secondary'}}>-{Math.round(windowMs / 60000)} min</Typography>
                <Typography variant="caption" sx={{color: 'text.secondary'}}>scale 0 - {format(max)}</Typography>
                <Typography variant="caption" sx={{color: 'text.secondary'}}>now</Typography>
            </Box>
        </Box>
    );
}
