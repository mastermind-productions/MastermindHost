import {useEffect, useState} from 'react';
import {Button, IconButton, Paper, Stack, Table, TableBody, TableCell, TableHead, TableRow, TextField, Typography} from '@mui/material';
import DeleteIcon from '@mui/icons-material/Delete';
import AddIcon from '@mui/icons-material/Add';
import {backend, Host, run} from '../../api';
import type {TabProps} from '../ServerPage';

export default function ScriptConfigTab({server: s, onChanged}: TabProps) {
    const [spec, setSpec] = useState<backend.ScriptSpec | null>(null);
    useEffect(() => { Host.GetScript(s.id).then(setSpec); }, [s.id]);
    if (!spec) return null;

    const set = (patch: Partial<backend.ScriptSpec>) => setSpec(new backend.ScriptSpec({...spec, ...patch}));
    const setProc = (i: number, patch: Partial<backend.ScriptProc>) =>
        set({procs: spec.procs.map((p, j) => j === i ? new backend.ScriptProc({...p, ...patch}) : p)});

    return (
        <Stack spacing={2}>
            <Paper sx={{p: 2}}>
                <Stack spacing={2}>
                    <TextField label="Working folder" value={spec.workDir} onChange={e => set({workDir: e.target.value})}/>
                    <TextField label="Start command (blank = launch the executables below in order)" value={spec.startCmd} onChange={e => set({startCmd: e.target.value})}
                               helperText='Run hidden through cmd, e.g. "F:\MMRappelz\Start-Servers.bat" -NoPause'/>
                    <TextField label="Stop command (optional - blank = close each component gracefully in reverse order, then force-stop)" value={spec.stopCmd} onChange={e => set({stopCmd: e.target.value})}/>
                </Stack>
            </Paper>
            <Paper>
                <Stack direction="row" sx={{p: 1.5, alignItems: 'center'}}>
                    <Typography variant="subtitle1" sx={{fontWeight: 600, flex: 1}}>Components (start order)</Typography>
                    <Button startIcon={<AddIcon/>} onClick={() => set({procs: [...spec.procs, new backend.ScriptProc({name: '', exe: '', port: 0})]})}>Add</Button>
                </Stack>
                <Table size="small">
                    <TableHead><TableRow><TableCell>Name</TableCell><TableCell>Executable (full path or relative to working folder)</TableCell><TableCell width={110}>Port</TableCell><TableCell width={50}/></TableRow></TableHead>
                    <TableBody>
                        {spec.procs.map((p, i) => (
                            <TableRow key={i}>
                                <TableCell><TextField size="small" value={p.name} onChange={e => setProc(i, {name: e.target.value})}/></TableCell>
                                <TableCell><TextField size="small" fullWidth value={p.exe} onChange={e => setProc(i, {exe: e.target.value})}/></TableCell>
                                <TableCell><TextField size="small" type="number" value={p.port} onChange={e => setProc(i, {port: Number(e.target.value) || 0})}/></TableCell>
                                <TableCell><IconButton onClick={() => set({procs: spec.procs.filter((_, j) => j !== i)})}><DeleteIcon/></IconButton></TableCell>
                            </TableRow>
                        ))}
                    </TableBody>
                </Table>
            </Paper>
            <Stack direction="row">
                <Button variant="contained" onClick={() => run(Host.SetScript(s.id, spec), 'Saved').then(onChanged)}>Save</Button>
            </Stack>
            <Typography variant="body2" sx={{color: 'text.secondary'}}>
                Components are used to detect whether the server is up (matched by executable path) and, when listed with a port, whether it is listening. The server shows Online once every component is running and every listed port is listening.
            </Typography>
        </Stack>
    );
}
