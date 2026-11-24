import {useEffect, useState} from 'react';
import {Button, Chip, Paper, Stack, Table, TableBody, TableCell, TableHead, TableRow, Typography} from '@mui/material';
import {backend, EventsOn, Host, run} from '../../api';
import type {TabProps} from '../ServerPage';

export default function EventsTab({server}: TabProps) {
    const [events, setEvents] = useState<backend.Event[]>([]);
    const load = () => Host.GetEvents(server.id).then(setEvents);
    useEffect(() => {
        load();
        return EventsOn('gs-event', (id: string) => { if (id === server.id) load(); });
    }, [server.id]);

    return (
        <Paper>
            <Stack direction="row" sx={{p: 1.5, alignItems: 'center'}}>
                <Typography variant="subtitle1" sx={{fontWeight: 600, flex: 1}}>Event history</Typography>
                <Button color="error" onClick={() => run(Host.ClearEvents(server.id), 'Cleared').then(load)}>Clear</Button>
            </Stack>
            <Table size="small">
                <TableHead><TableRow><TableCell>Time</TableCell><TableCell>Action</TableCell><TableCell>Result</TableCell><TableCell>Message</TableCell></TableRow></TableHead>
                <TableBody>
                    {events.map((e, i) => (
                        <TableRow key={i}>
                            <TableCell sx={{whiteSpace: 'nowrap'}}>{new Date(e.time).toLocaleString()}</TableCell>
                            <TableCell>{e.action}</TableCell>
                            <TableCell><Chip size="small" label={e.result} color={e.result === 'success' ? 'success' : e.result === 'error' ? 'error' : 'default'}/></TableCell>
                            <TableCell sx={{wordBreak: 'break-word'}}>{e.message}</TableCell>
                        </TableRow>
                    ))}
                    {!events.length && <TableRow><TableCell colSpan={4} sx={{color: 'text.secondary'}}>No events.</TableCell></TableRow>}
                </TableBody>
            </Table>
        </Paper>
    );
}
