import {useEffect, useRef, useState} from 'react';
import {Box, Button, Paper, Stack, Table, TableBody, TableCell, TableHead, TableRow, TextField, Typography} from '@mui/material';
import RefreshIcon from '@mui/icons-material/Refresh';
import {backend, Host, isActive, run} from '../../api';
import type {TabProps} from '../ServerPage';

export default function PlayersTab({server}: TabProps) {
    const [players, setPlayers] = useState<backend.Player[]>([]);
    const [err, setErr] = useState('');
    const [cmd, setCmd] = useState('');
    const [msg, setMsg] = useState('');
    const [history, setHistory] = useState<string[]>([]);
    const consoleRef = useRef<HTMLPreElement>(null);
    const online = isActive(server.status);

    const load = async () => {
        if (!online) { setPlayers([]); setErr('Server is not running.'); return; }
        try { setPlayers(await Host.Players(server.id)); setErr(''); } catch (e: any) { setErr(String(e)); }
    };
    useEffect(() => { load(); const t = setInterval(load, 15000); return () => clearInterval(t); }, [server.id, online]);
    useEffect(() => { consoleRef.current?.scrollTo(0, consoleRef.current.scrollHeight); }, [history]);

    const exec = async (c: string) => {
        if (!c.trim()) return;
        setHistory(h => [...h, '> ' + c]);
        try { const out = await Host.RconCommand(server.id, c); setHistory(h => [...h, out]); } catch (e: any) { setHistory(h => [...h, 'ERROR: ' + e]); }
    };

    return (
        <Stack spacing={2}>
            <Paper>
                <Stack direction="row" sx={{alignItems: 'center', p: 1.5}}>
                    <Typography variant="subtitle1" sx={{fontWeight: 600, flex: 1}}>Online players ({players.length})</Typography>
                    <Button startIcon={<RefreshIcon/>} onClick={load}>Refresh</Button>
                </Stack>
                <Table size="small">
                    <TableHead><TableRow><TableCell>#</TableCell><TableCell>Name</TableCell><TableCell>EOS ID</TableCell><TableCell align="right">Actions</TableCell></TableRow></TableHead>
                    <TableBody>
                        {players.map(p => (
                            <TableRow key={p.id}>
                                <TableCell>{p.index}</TableCell>
                                <TableCell>{p.name}</TableCell>
                                <TableCell sx={{fontFamily: 'monospace'}}>{p.id}</TableCell>
                                <TableCell align="right">
                                    <Button onClick={() => navigator.clipboard.writeText(p.id)}>Copy ID</Button>
                                    <Button color="warning" onClick={() => run(Host.PlayerKick(server.id, p.id), 'Kicked').then(load)}>Kick</Button>
                                    <Button color="error" onClick={() => run(Host.PlayerBan(server.id, p.id), 'Banned').then(load)}>Ban</Button>
                                </TableCell>
                            </TableRow>
                        ))}
                        {!players.length && <TableRow><TableCell colSpan={4} sx={{color: 'text.secondary'}}>{err || 'No players online.'}</TableCell></TableRow>}
                    </TableBody>
                </Table>
            </Paper>
            <Paper sx={{p: 2}}>
                <Typography variant="subtitle1" sx={{fontWeight: 600, mb: 1}}>Broadcast</Typography>
                <Stack direction="row" spacing={1}>
                    <TextField fullWidth label="Message to all players" value={msg} onChange={e => setMsg(e.target.value)}
                               onKeyDown={e => { if (e.key === 'Enter' && msg) { run(Host.Broadcast(server.id, msg), 'Sent'); setMsg(''); } }}/>
                    <Button variant="contained" disabled={!msg || !online} onClick={() => { run(Host.Broadcast(server.id, msg), 'Sent'); setMsg(''); }}>Send</Button>
                </Stack>
            </Paper>
            <Paper sx={{p: 2}}>
                <Typography variant="subtitle1" sx={{fontWeight: 600, mb: 1}}>RCON console</Typography>
                <Box component="pre" ref={consoleRef} sx={{m: 0, mb: 1, p: 1, height: 260, overflow: 'auto', bgcolor: '#000', borderRadius: 1, fontSize: 12, fontFamily: 'Consolas, monospace', color: '#8fe3a8', whiteSpace: 'pre-wrap'}}>
                    {history.join('\n') || 'Examples: SaveWorld, ListPlayers, DestroyWildDinos, SetTimeOfDay 08:00, GetChat'}
                </Box>
                <Stack direction="row" spacing={1}>
                    {['SaveWorld', 'DestroyWildDinos', 'GetChat'].map(c => <Button key={c} variant="outlined" disabled={!online} onClick={() => exec(c)}>{c}</Button>)}
                    <TextField fullWidth placeholder="Command" value={cmd} onChange={e => setCmd(e.target.value)}
                               onKeyDown={e => { if (e.key === 'Enter') { exec(cmd); setCmd(''); } }}/>
                    <Button variant="contained" disabled={!online} onClick={() => { exec(cmd); setCmd(''); }}>Run</Button>
                </Stack>
            </Paper>
        </Stack>
    );
}
