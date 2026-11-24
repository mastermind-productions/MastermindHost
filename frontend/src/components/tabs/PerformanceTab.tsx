import {useCallback, useEffect, useState} from 'react';
import {Alert, Box, Button, Chip, Paper, Stack, ToggleButton, ToggleButtonGroup, Typography} from '@mui/material';
import {backend, fmtBytes, Host, run} from '../../api';
import LineChart, {Pt} from '../charts/LineChart';
import type {TabProps} from '../ServerPage';

const RANGES: [string, number][] = [['5 min', 300], ['15 min', 900], ['1 hour', 3600]];
const pct = (v: number) => `${v.toFixed(1)}%`;

export default function PerformanceTab({server, onChanged}: TabProps) {
    const [range, setRange] = useState(300);
    const [sv, setSv] = useState<backend.MetricPoint[]>([]);
    const [host, setHost] = useState<{points: backend.MetricPoint[]; ramTotal: number; cores: number} | null>(null);
    const [aff, setAff] = useState<backend.AffinityInfo | null>(null);
    const [sel, setSel] = useState<number[]>([]);

    const loadAff = useCallback(() => Host.GetAffinity(server.id).then(a => {
        setAff(a);
        setSel(a.saved.length ? a.saved : Array.from({length: a.cores}, (_, i) => i));
    }), [server.id]);

    useEffect(() => { loadAff(); }, [loadAff]);
    useEffect(() => {
        const load = () => {
            Host.ServerMetrics(server.id, range).then(setSv);
            Host.HostMetrics(range).then(h => setHost(h as any));
        };
        load();
        const t = setInterval(load, 2000);
        return () => clearInterval(t);
    }, [server.id, range]);

    const toPts = (m: backend.MetricPoint[] | undefined, f: (p: backend.MetricPoint) => number): Pt[] => (m ?? []).map(p => ({t: p.t, v: f(p)}));
    const cores = aff?.cores ?? 0;
    const toggle = (i: number) => setSel(s => s.includes(i) ? s.filter(x => x !== i) : [...s, i].sort((a, b) => a - b));
    const allSel = cores > 0 && sel.length === cores;
    const windowMs = range * 1000;
    const peakRam = Math.max(1, ...sv.map(p => p.ram));

    return (
        <Stack spacing={2}>
            <Stack direction="row" spacing={2} sx={{alignItems: 'center'}}>
                <Typography variant="subtitle1" sx={{fontWeight: 600, flex: 1}}>Resource history</Typography>
                <ToggleButtonGroup size="small" exclusive value={range} onChange={(_, v) => v && setRange(v)}>
                    {RANGES.map(([l, v]) => <ToggleButton key={v} value={v}>{l}</ToggleButton>)}
                </ToggleButtonGroup>
            </Stack>
            {!server.startTime && <Alert severity="info">Server is not running. The server graphs fill in while it runs; host graphs are always recorded while MastermindHost is open.</Alert>}
            <Box sx={{display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(380px, 1fr))', gap: 2}}>
                <LineChart title="Server CPU (% of all cores)" points={toPts(sv, p => p.cpu)} max={100} format={pct} windowMs={windowMs}/>
                <LineChart title="Server RAM" points={toPts(sv, p => p.ram)} max={Math.max(peakRam * 1.25, 1 << 30)} format={fmtBytes} windowMs={windowMs}/>
                <LineChart title="Host CPU" points={toPts(host?.points, p => p.cpu)} max={100} format={pct} color="#38bdf8" windowMs={windowMs}/>
                <LineChart title="Host RAM in use" points={toPts(host?.points, p => p.ram)} max={host?.ramTotal || 1} format={fmtBytes} color="#38bdf8" windowMs={windowMs}/>
            </Box>
            <Paper sx={{p: 2}}>
                <Stack direction="row" spacing={1} sx={{alignItems: 'center', mb: 1}}>
                    <Typography variant="subtitle1" sx={{fontWeight: 600}}>CPU affinity</Typography>
                    {aff?.running && <Chip size="small" label={`Live: ${aff.current.length}/${aff.cores} CPUs`}/>}
                    {aff && aff.saved.length > 0 && <Chip size="small" color="success" label={`Pinned: ${aff.saved.length} CPUs`}/>}
                </Stack>
                <Typography variant="body2" sx={{color: 'text.secondary', mb: 2}}>
                    Choose the logical CPUs this server may use. Applies immediately if running and on every start. Select all to remove the limit.
                </Typography>
                <Box sx={{display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(70px, 1fr))', gap: 1, mb: 2}}>
                    {Array.from({length: cores}, (_, i) => (
                        <Button key={i} size="small" variant={sel.includes(i) ? 'contained' : 'outlined'} onClick={() => toggle(i)}>CPU {i}</Button>
                    ))}
                </Box>
                <Stack direction="row" spacing={1}>
                    <Button variant="contained" disabled={!sel.length}
                            onClick={() => run(Host.SetAffinity(server.id, allSel ? [] : sel), 'Affinity saved').then(() => { loadAff(); onChanged(); })}>Apply</Button>
                    <Button onClick={() => setSel(Array.from({length: cores}, (_, i) => i))}>All</Button>
                    <Button onClick={() => setSel(Array.from({length: cores / 2}, (_, i) => i * 2))}>Even only</Button>
                    <Button onClick={() => setSel(Array.from({length: Math.floor(cores / 2)}, (_, i) => i))}>First half</Button>
                    <Button onClick={() => setSel(Array.from({length: cores - Math.floor(cores / 2)}, (_, i) => i + Math.floor(cores / 2)))}>Second half</Button>
                </Stack>
            </Paper>
        </Stack>
    );
}
