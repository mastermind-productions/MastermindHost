import {useEffect, useState} from 'react';
import {Box, Button, Checkbox, Dialog, DialogActions, DialogContent, DialogTitle, FormControlLabel, MenuItem, Paper, Stack, Switch, Table, TableBody, TableCell, TableHead, TableRow, TextField, ToggleButton, ToggleButtonGroup, Typography} from '@mui/material';
import {backend, Host, run} from '../../api';
import type {TabProps} from '../ServerPage';

const ACTIONS = ['restart', 'shutdown', 'start', 'stop', 'backup', 'update', 'broadcast', 'rcon'];
const DAYS = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat'];

const blank = () => new backend.Schedule({id: '', enabled: true, action: 'restart', mode: 'daily', time: '04:00', days: [], intervalMin: 60, arg: '', warnMinutes: 15});

const describe = (s: backend.Schedule) =>
    s.mode === 'interval' ? `Every ${s.intervalMin} min` : `${s.days?.length ? s.days.map(d => DAYS[d]).join(', ') : 'Daily'} at ${s.time}`;

export default function SchedulerTab({server}: TabProps) {
    const [list, setList] = useState<backend.Schedule[]>([]);
    const [edit, setEdit] = useState<backend.Schedule | null>(null);

    const load = () => Host.GetSchedules(server.id).then(setList);
    useEffect(() => { load(); }, [server.id]);

    const save = async (s: backend.Schedule) => {
        if (await run(Host.SaveSchedule(server.id, s), 'Schedule saved')) { setEdit(null); load(); }
    };
    const set = (p: Partial<backend.Schedule>) => edit && setEdit(new backend.Schedule({...edit, ...p}));
    const needsArg = edit && (edit.action === 'broadcast' || edit.action === 'rcon');
    const warnable = edit && ['restart', 'shutdown', 'stop', 'update'].includes(edit.action);

    return (
        <Stack spacing={2}>
            <Paper>
                <Stack direction="row" sx={{p: 1.5, alignItems: 'center'}}>
                    <Typography variant="subtitle1" sx={{fontWeight: 600, flex: 1}}>Scheduled tasks</Typography>
                    <Button variant="contained" onClick={() => setEdit(blank())}>Add task</Button>
                </Stack>
                <Table size="small">
                    <TableHead><TableRow><TableCell>On</TableCell><TableCell>Action</TableCell><TableCell>When</TableCell><TableCell>Warn</TableCell><TableCell>Last run</TableCell><TableCell align="right">Actions</TableCell></TableRow></TableHead>
                    <TableBody>
                        {list.map(s => (
                            <TableRow key={s.id}>
                                <TableCell><Switch size="small" checked={s.enabled} onChange={e => save(new backend.Schedule({...s, enabled: e.target.checked}))}/></TableCell>
                                <TableCell sx={{textTransform: 'capitalize'}}>{s.action}{s.arg ? `: ${s.arg}` : ''}</TableCell>
                                <TableCell>{describe(s)}</TableCell>
                                <TableCell>{s.warnMinutes ? `${s.warnMinutes} min` : '-'}</TableCell>
                                <TableCell>{s.lastRun && !String(s.lastRun).startsWith('0001') ? new Date(s.lastRun).toLocaleString() : '-'}</TableCell>
                                <TableCell align="right" sx={{whiteSpace: 'nowrap'}}>
                                    <Button onClick={() => run(Host.RunScheduleNow(server.id, s.id), 'Task started')}>Run now</Button>
                                    <Button onClick={() => setEdit(s)}>Edit</Button>
                                    <Button color="error" onClick={() => run(Host.DeleteSchedule(server.id, s.id), 'Deleted').then(load)}>Delete</Button>
                                </TableCell>
                            </TableRow>
                        ))}
                        {!list.length && <TableRow><TableCell colSpan={6} sx={{color: 'text.secondary'}}>No scheduled tasks.</TableCell></TableRow>}
                    </TableBody>
                </Table>
            </Paper>
            <Dialog open={!!edit} onClose={() => setEdit(null)} fullWidth maxWidth="sm">
                <DialogTitle>{edit?.id ? 'Edit task' : 'New task'}</DialogTitle>
                {edit && (
                    <DialogContent>
                        <Stack spacing={2} sx={{mt: 1}}>
                            <TextField select label="Action" value={edit.action} onChange={e => set({action: e.target.value})}>
                                {ACTIONS.filter(a => !server.script || ['restart', 'shutdown', 'start', 'stop'].includes(a)).map(a => <MenuItem key={a} value={a} sx={{textTransform: 'capitalize'}}>{a}</MenuItem>)}
                            </TextField>
                            <ToggleButtonGroup exclusive size="small" value={edit.mode} onChange={(_, v) => v && set({mode: v})}>
                                <ToggleButton value="daily">At a time of day</ToggleButton>
                                <ToggleButton value="interval">Every N minutes</ToggleButton>
                            </ToggleButtonGroup>
                            {edit.mode === 'daily' ? (
                                <>
                                    <TextField label="Time (24h)" type="time" value={edit.time} onChange={e => set({time: e.target.value})} sx={{width: 180}}/>
                                    <Box>
                                        <Typography variant="caption" sx={{color: 'text.secondary'}}>Days (none = every day)</Typography>
                                        <Box>
                                            {DAYS.map((d, i) => (
                                                <FormControlLabel key={d} label={d} control={<Checkbox size="small" checked={edit.days?.includes(i) ?? false}
                                                    onChange={e => set({days: e.target.checked ? [...(edit.days ?? []), i].sort() : (edit.days ?? []).filter(x => x !== i)})}/>}/>
                                            ))}
                                        </Box>
                                    </Box>
                                </>
                            ) : (
                                <TextField label="Interval (minutes)" type="number" value={edit.intervalMin} onChange={e => set({intervalMin: Number(e.target.value)})} sx={{width: 180}}/>
                            )}
                            {warnable && (
                                <TextField label="Warn players (minutes before, 0 = none)" type="number" value={edit.warnMinutes} onChange={e => set({warnMinutes: Number(e.target.value)})}/>
                            )}
                            {(needsArg || warnable) && (
                                <TextField label={edit.action === 'rcon' ? 'RCON command' : edit.action === 'broadcast' ? 'Message' : 'Warning message (%s = action, %d = minutes)'}
                                           placeholder={warnable ? 'Server %s in %d minute(s).' : ''} value={edit.arg} onChange={e => set({arg: e.target.value})}/>
                            )}
                            <FormControlLabel label="Enabled" control={<Switch checked={edit.enabled} onChange={e => set({enabled: e.target.checked})}/>}/>
                        </Stack>
                    </DialogContent>
                )}
                <DialogActions>
                    <Button onClick={() => setEdit(null)}>Cancel</Button>
                    <Button variant="contained" onClick={() => edit && save(edit)}>Save</Button>
                </DialogActions>
            </Dialog>
        </Stack>
    );
}
