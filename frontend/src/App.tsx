import {useCallback, useEffect, useRef, useState} from 'react';
import {Alert, AppBar, Box, Button, Collapse, CssBaseline, IconButton, Snackbar, ThemeProvider, Toolbar, Tooltip, Typography} from '@mui/material';
import ExpandMoreIcon from '@mui/icons-material/ExpandMore';
import AddIcon from '@mui/icons-material/Add';
import SettingsIcon from '@mui/icons-material/Settings';
import RefreshIcon from '@mui/icons-material/Refresh';
import DnsIcon from '@mui/icons-material/Dns';
import {theme} from './theme';
import {backend, EventsOn, fmtBytes, Host, run, setNotifier} from './api';
import ServerCard from './components/ServerCard';
import CreateDialog from './components/CreateDialog';
import SettingsDialog from './components/SettingsDialog';
import ServerPage from './components/ServerPage';

const GROUPS = ['ASA', 'ASE', 'Palworld', 'MMOs'];

type Toast = {msg: string; severity: 'success' | 'error' | 'info'; key: number};

export default function App() {
    const [servers, setServers] = useState<backend.ServerInfo[]>([]);
    const [stats, setStats] = useState<Record<string, any>>({});
    const [selected, setSelected] = useState<string | null>(null);
    const [createOpen, setCreateOpen] = useState(false);
    const [settingsOpen, setSettingsOpen] = useState(false);
    const [toast, setToast] = useState<Toast | null>(null);
    const pending = useRef<number | null>(null);
    const [collapsed, setCollapsed] = useState<Record<string, boolean>>({});
    const [dragId, setDragId] = useState<string | null>(null);

    useEffect(() => setNotifier((msg, severity) => setToast({msg, severity, key: Date.now()})), []);

    const refresh = useCallback(async () => {
        const list = await run(Host.List());
        if (list) setServers(list);
    }, []);

    useEffect(() => {
        refresh();
        const off = EventsOn('gs-status-update', () => {
            if (pending.current) return;
            pending.current = window.setTimeout(() => { pending.current = null; refresh(); }, 400);
        });
        const loadStats = () => Host.HostStats().then(setStats).catch(() => {});
        loadStats();
        const t = window.setInterval(() => { loadStats(); refresh(); }, 5000);
        return () => { off(); window.clearInterval(t); };
    }, [refresh]);

    const current = servers.find(s => s.id === selected);

    const grouped = [...GROUPS, ...Array.from(new Set(servers.map(s => s.group))).filter(g => !GROUPS.includes(g))]
        .map(g => ({name: g, items: servers.filter(s => s.group === g)})).filter(g => g.items.length > 0);

    const reorder = async (items: backend.ServerInfo[], from: number, to: number) => {
        if (from === to || from < 0 || to < 0 || to >= items.length) return;
        const ids = items.map(s => s.id);
        const [moved] = ids.splice(from, 1);
        ids.splice(to, 0, moved);
        const order = new Map<string, number>();
        servers.forEach((s, i) => order.set(s.id, i));
        const slots = ids.map(id => order.get(id)!).sort((a, b) => a - b);
        const next = [...servers];
        ids.forEach((id, n) => { next[slots[n]] = servers.find(s => s.id === id)!; });
        setServers(next);
        await run(Host.ReorderServers(ids));
        refresh();
    };

    return (
        <ThemeProvider theme={theme}>
            <CssBaseline/>
            <Box sx={{height: '100vh', display: 'flex', flexDirection: 'column'}}>
                <AppBar position="static" elevation={0}>
                    <Toolbar variant="dense" sx={{gap: 2}}>
                        <DnsIcon color="primary"/>
                        <Typography variant="h6" sx={{cursor: 'pointer', fontWeight: 700}} onClick={() => setSelected(null)}>
                            <Box component="span" sx={{color: 'primary.main'}}>Mastermind</Box>Host
                        </Typography>
                        <Box sx={{flex: 1}}/>
                        <Stat label="CPU" value={`${(stats.cpu ?? 0).toFixed(0)}%`}/>
                        <Stat label="RAM" value={`${fmtBytes(stats.ramUsed)} / ${fmtBytes(stats.ramTotal)}`}/>
                        <Stat label="Disk" value={`${fmtBytes(stats.diskUsed)} / ${fmtBytes(stats.diskTotal)}`}/>
                        <Tooltip title="Check for game updates"><IconButton onClick={() => run(Host.CheckForUpdates(), 'Update check complete')}><RefreshIcon/></IconButton></Tooltip>
                        <Tooltip title="Settings"><IconButton onClick={() => setSettingsOpen(true)}><SettingsIcon/></IconButton></Tooltip>
                    </Toolbar>
                </AppBar>
                <Box sx={{flex: 1, overflow: 'auto', p: 2}}>
                    {current ? (
                        <ServerPage server={current} onBack={() => setSelected(null)} onChanged={refresh}/>
                    ) : (
                        <>
                            <Box sx={{display: 'flex', alignItems: 'center', mb: 2}}>
                                <Typography variant="h5" sx={{fontWeight: 700}}>Game Servers</Typography>
                                <Typography sx={{ml: 1.5, color: 'text.secondary'}}>{servers.length} total</Typography>
                                <Box sx={{flex: 1}}/>
                                <Button variant="contained" startIcon={<AddIcon/>} onClick={() => setCreateOpen(true)}>Create Server</Button>
                            </Box>
                            {servers.length === 0 && (
                                <Box sx={{textAlign: 'center', mt: 10, color: 'text.secondary'}}>
                                    <Typography variant="h6">No servers yet</Typography>
                                    <Typography>Create a server to get started.</Typography>
                                </Box>
                            )}
                            {grouped.map(g => (
                                <Box key={g.name} sx={{mb: 3}}>
                                    <Box sx={{display: 'flex', alignItems: 'center', mb: 1, cursor: 'pointer', userSelect: 'none'}}
                                         onClick={() => setCollapsed(c => ({...c, [g.name]: !c[g.name]}))}>
                                        <ExpandMoreIcon sx={{transform: collapsed[g.name] ? 'rotate(-90deg)' : 'none', transition: 'transform .15s', color: 'primary.main'}}/>
                                        <Typography variant="h6" sx={{fontWeight: 700, ml: 0.5}}>{g.name}</Typography>
                                        <Typography sx={{ml: 1.5, color: 'text.secondary'}}>{g.items.length}</Typography>
                                    </Box>
                                    <Collapse in={!collapsed[g.name]} unmountOnExit={false}>
                                        <Box sx={{display: 'grid', gap: 2, gridTemplateColumns: 'repeat(auto-fill, minmax(360px, 1fr))'}}>
                                            {g.items.map((s, i) => (
                                                <ServerCard key={s.id} server={s} onConfigure={() => setSelected(s.id)} onChanged={refresh}
                                                            order={{
                                                                onMove: d => reorder(g.items, i, i + d),
                                                                canPrev: i > 0, canNext: i < g.items.length - 1,
                                                                dragging: dragId === s.id,
                                                                onDragStart: () => setDragId(s.id),
                                                                onDragEnd: () => setDragId(null),
                                                                onDropHere: () => {
                                                                    const from = g.items.findIndex(x => x.id === dragId);
                                                                    setDragId(null);
                                                                    if (from >= 0) reorder(g.items, from, i);
                                                                },
                                                            }}/>
                                            ))}
                                        </Box>
                                    </Collapse>
                                </Box>
                            ))}
                        </>
                    )}
                </Box>
            </Box>
            <CreateDialog open={createOpen} onClose={() => setCreateOpen(false)} onCreated={id => { setCreateOpen(false); refresh(); setSelected(id); }}/>
            <SettingsDialog open={settingsOpen} onClose={() => setSettingsOpen(false)}/>
            <Snackbar key={toast?.key} open={!!toast} autoHideDuration={toast?.severity === 'error' ? 8000 : 3000}
                      onClose={() => setToast(null)} anchorOrigin={{vertical: 'bottom', horizontal: 'right'}}>
                <Alert severity={toast?.severity ?? 'info'} variant="filled" onClose={() => setToast(null)} sx={{whiteSpace: 'pre-wrap', maxWidth: 600}}>
                    {toast?.msg}
                </Alert>
            </Snackbar>
        </ThemeProvider>
    );
}

function Stat({label, value}: {label: string; value: string}) {
    return (
        <Box sx={{textAlign: 'right', lineHeight: 1.1}}>
            <Typography variant="caption" sx={{color: 'text.secondary', display: 'block'}}>{label}</Typography>
            <Typography variant="body2" sx={{fontWeight: 600}}>{value}</Typography>
        </Box>
    );
}
