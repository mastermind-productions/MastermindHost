import {useEffect, useState} from 'react';
import {Box, Button, Chip, Dialog, DialogActions, DialogContent, DialogContentText, DialogTitle, Paper, Stack, Table, TableBody, TableCell, TableHead, TableRow, TextField, Typography} from '@mui/material';
import FolderOpenIcon from '@mui/icons-material/FolderOpen';
import {backend, fmtBytes, fmtUptime, Host, run} from '../../api';
import {Field} from '../ServerCard';
import type {TabProps} from '../ServerPage';

export default function ScriptOverviewTab({server: s, onChanged}: TabProps) {
    const [name, setName] = useState(s.name);
    const [comps, setComps] = useState<backend.ScriptComponent[]>([]);
    const [confirm, setConfirm] = useState(false);

    useEffect(() => {
        let alive = true;
        const load = () => Host.ScriptComponents(s.id).then(c => { if (alive) setComps(c); }).catch(() => {});
        load();
        const t = window.setInterval(load, 3000);
        return () => { alive = false; window.clearInterval(t); };
    }, [s.id]);

    return (
        <Stack spacing={2}>
            <Paper sx={{p: 2}}>
                <Box sx={{display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(150px, 1fr))', gap: 2}}>
                    <Field label="Components up" value={`${s.compUp} / ${s.compTotal}`}/>
                    <Field label="Game port" value={s.port ? String(s.port) : '-'}/>
                    <Field label="CPU (all components)" value={s.startTime ? `${s.cpu.toFixed(1)}%` : '-'}/>
                    <Field label="RAM (all components)" value={s.startTime ? fmtBytes(s.ram) : '-'}/>
                    <Field label="Uptime" value={fmtUptime(s.startTime)}/>
                    <Field label="Added" value={new Date(s.createdAt).toLocaleString()}/>
                </Box>
            </Paper>
            <Paper>
                <Typography variant="subtitle1" sx={{fontWeight: 600, p: 1.5}}>Components</Typography>
                <Table size="small">
                    <TableHead>
                        <TableRow><TableCell>Name</TableCell><TableCell>Port</TableCell><TableCell>State</TableCell><TableCell>PID</TableCell><TableCell>Executable</TableCell></TableRow>
                    </TableHead>
                    <TableBody>
                        {comps.map(c => (
                            <TableRow key={c.exe}>
                                <TableCell>{c.name}</TableCell>
                                <TableCell>{c.port || '-'}</TableCell>
                                <TableCell>
                                    <Chip size="small" color={c.running ? (c.port && !c.listening ? 'warning' : 'success') : 'default'}
                                          label={c.running ? (c.port ? (c.listening ? 'listening' : 'running, not listening') : 'running') : 'stopped'}/>
                                </TableCell>
                                <TableCell>{c.pid || '-'}</TableCell>
                                <TableCell sx={{wordBreak: 'break-all'}}>{c.exe}</TableCell>
                            </TableRow>
                        ))}
                        {comps.length === 0 && <TableRow><TableCell colSpan={5} sx={{color: 'text.secondary'}}>No components defined - add executables in Launch Config.</TableCell></TableRow>}
                    </TableBody>
                </Table>
            </Paper>
            <Paper sx={{p: 2}}>
                <Stack spacing={2}>
                    <Stack direction="row" spacing={1}>
                        <TextField label="Display name (the folder is not renamed)" value={name} onChange={e => setName(e.target.value)} sx={{flex: 1}}/>
                        <Button variant="contained" disabled={!name.trim() || name === s.name}
                                onClick={() => run(Host.SetName(s.id, name), 'Renamed').then(onChanged)}>Rename</Button>
                    </Stack>
                    <Stack direction="row" spacing={1}>
                        <Button variant="outlined" startIcon={<FolderOpenIcon/>} onClick={() => run(Host.OpenFolder(s.id, ''))}>Server folder</Button>
                        <Box sx={{flex: 1}}/>
                        <Button color="error" variant="outlined" onClick={() => setConfirm(true)}>Remove from list</Button>
                    </Stack>
                </Stack>
            </Paper>
            <Dialog open={confirm} onClose={() => setConfirm(false)}>
                <DialogTitle>Remove server</DialogTitle>
                <DialogContent><DialogContentText>Remove "{s.name}" from MastermindHost? Files in {s.dir} are kept.</DialogContentText></DialogContent>
                <DialogActions>
                    <Button onClick={() => setConfirm(false)}>Cancel</Button>
                    <Button color="error" variant="contained" onClick={async () => {
                        await run(Host.Remove(s.id), 'Server removed');
                        setConfirm(false); onChanged();
                    }}>Remove</Button>
                </DialogActions>
            </Dialog>
        </Stack>
    );
}
