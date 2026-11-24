import {useCallback, useEffect, useState} from 'react';
import {Alert, Button, Chip, Paper, Stack, Table, TableBody, TableCell, TableHead, TableRow, TextField, Typography} from '@mui/material';
import {backend, Host, run} from '../../api';
import type {TabProps} from '../ServerPage';

const fmtDur = (s: number) => `${Math.floor(s / 3600)}h ${Math.floor((s % 3600) / 60)}m`;

export default function NetworkTab({server}: TabProps) {
    const [report, setReport] = useState<backend.ServerPorts[]>([]);
    const [a2s, setA2s] = useState<backend.A2SResult | null>(null);
    const [a2sErr, setA2sErr] = useState('');
    const [addr, setAddr] = useState('');
    const [port, setPort] = useState('');
    const [free, setFree] = useState('');
    const [busy, setBusy] = useState(false);

    const load = useCallback(() => { run(Host.PortsReport()).then(r => r && setReport(r)); }, []);
    useEffect(() => { load(); }, [load]);

    const query = (p: Promise<backend.A2SResult>) => {
        setBusy(true); setA2sErr('');
        p.then(setA2s).catch(e => { setA2s(null); setA2sErr(String(e?.message ?? e)); }).finally(() => setBusy(false));
    };
    const mine = report.find(r => r.id === server.id);
    const conflictCount = report.reduce((n, r) => n + r.ports.filter(p => p.conflicts).length, 0);

    return (
        <Stack spacing={2}>
            <Paper sx={{p: 2}}>
                <Stack direction="row" spacing={1} sx={{alignItems: 'center', mb: 1}}>
                    <Typography variant="subtitle1" sx={{fontWeight: 600, flex: 1}}>Ports manager</Typography>
                    {conflictCount > 0 ? <Chip color="warning" size="small" label={`${conflictCount} conflict(s)`}/> : <Chip color="success" size="small" label="No conflicts"/>}
                    <Button onClick={load}>Rescan</Button>
                </Stack>
                {report.length === 0 && <Typography variant="body2" sx={{color: 'text.secondary'}}>No data yet.</Typography>}
                {report.map(r => (
                    <Stack key={r.id} sx={{mb: 2, opacity: r.id === server.id ? 1 : 0.8}}>
                        <Stack direction="row" spacing={1} sx={{alignItems: 'center'}}>
                            <Typography sx={{fontWeight: 600}}>{r.name}{r.id === server.id ? ' (this server)' : ''}</Typography>
                            <Chip size="small" label={r.gameName}/>
                            <Chip size="small" color={r.running ? 'success' : 'default'} label={r.running ? 'running' : 'stopped'}/>
                            <Chip size="small" color={r.firewall ? 'success' : 'default'} label={r.firewall ? 'firewall rule' : 'no firewall rule'}/>
                        </Stack>
                        <Table size="small">
                            <TableHead><TableRow><TableCell>Purpose</TableCell><TableCell>Proto</TableCell><TableCell>Port</TableCell><TableCell>Listening</TableCell><TableCell>Notes</TableCell></TableRow></TableHead>
                            <TableBody>
                                {r.ports.map(p => (
                                    <TableRow key={p.label}>
                                        <TableCell>{p.label}</TableCell><TableCell>{p.proto}</TableCell><TableCell>{p.port}</TableCell>
                                        <TableCell>{p.bound ? (p.ours ? 'this server' : (p.process || `PID ${p.pid}`)) : '-'}</TableCell>
                                        <TableCell sx={{color: p.conflicts ? 'warning.main' : 'text.secondary'}}>{p.conflicts || 'ok'}</TableCell>
                                    </TableRow>
                                ))}
                            </TableBody>
                        </Table>
                    </Stack>
                ))}
                <Stack direction="row" spacing={1} sx={{alignItems: 'center'}}>
                    <Button variant="outlined" onClick={() => run(Host.FindFreePort('UDP', 7777)).then(p => p && setFree(`Next free UDP port from 7777: ${p}`))}>Find free UDP port</Button>
                    <Button variant="outlined" onClick={() => run(Host.FindFreePort('TCP', 27020)).then(p => p && setFree(`Next free TCP port from 27020: ${p}`))}>Find free TCP port</Button>
                    <Typography variant="body2">{free}</Typography>
                </Stack>
                <Typography variant="caption" sx={{color: 'text.secondary', display: 'block', mt: 1}}>
                    Detects local listeners and duplicate configuration only. Router port-forwarding cannot be verified from here; use Open ports under Discord / Beta / Firewall for Windows Firewall rules.
                </Typography>
            </Paper>
            <Paper sx={{p: 2}}>
                <Typography variant="subtitle1" sx={{fontWeight: 600, mb: 1}}>A2S query</Typography>
                <Typography variant="body2" sx={{color: 'text.secondary', mb: 2}}>
                    Steam server query (A2S_INFO / A2S_PLAYER) against the server's query port{mine ? ` (${mine.ports.find(p => p.label.startsWith('Query'))?.port})` : ''}, or any remote address.
                </Typography>
                <Stack direction="row" spacing={1} sx={{mb: 2}}>
                    <Button variant="contained" disabled={busy || !server.startTime} onClick={() => query(Host.QueryA2S(server.id))}>Query this server</Button>
                    <TextField size="small" label="Remote address" value={addr} onChange={e => setAddr(e.target.value)}/>
                    <TextField size="small" label="Query port" value={port} onChange={e => setPort(e.target.value.replace(/\D/g, ''))} sx={{width: 120}}/>
                    <Button variant="outlined" disabled={busy || !addr || !port} onClick={() => query(Host.QueryA2SAddress(addr, Number(port)))}>Query</Button>
                </Stack>
                {a2sErr && <Alert severity="warning">{a2sErr}</Alert>}
                {a2s && (
                    <Stack spacing={1}>
                        <Typography sx={{fontWeight: 600}}>{a2s.name}</Typography>
                        <Stack direction="row" spacing={1} sx={{flexWrap: 'wrap', rowGap: 1}}>
                            <Chip size="small" label={`Map: ${a2s.map}`}/>
                            <Chip size="small" label={`Players: ${a2s.players}/${a2s.maxPlayers}`}/>
                            <Chip size="small" label={`Ping: ${a2s.pingMs} ms`}/>
                            <Chip size="small" label={`Version: ${a2s.version || '-'}`}/>
                            <Chip size="small" label={a2s.passworded ? 'Password' : 'No password'}/>
                            <Chip size="small" label={`${a2s.serverType} / ${a2s.os}`}/>
                            <Chip size="small" label={`App ${a2s.appId}`}/>
                        </Stack>
                        {a2s.playerList.length > 0 && (
                            <Table size="small">
                                <TableHead><TableRow><TableCell>#</TableCell><TableCell>Name</TableCell><TableCell>Score</TableCell><TableCell>Connected</TableCell></TableRow></TableHead>
                                <TableBody>{a2s.playerList.map((p, i) => <TableRow key={i}><TableCell>{p.index}</TableCell><TableCell>{p.name || '(connecting)'}</TableCell><TableCell>{p.score}</TableCell><TableCell>{fmtDur(p.duration)}</TableCell></TableRow>)}</TableBody>
                            </Table>
                        )}
                    </Stack>
                )}
            </Paper>
        </Stack>
    );
}
