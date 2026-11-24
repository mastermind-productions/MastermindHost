import {useEffect, useMemo, useState} from 'react';
import {Box, Button, MenuItem, Paper, Stack, TextField, Typography} from '@mui/material';
import {backend, Host, run} from '../../api';
import type {TabProps} from '../ServerPage';

const SECRET = /password/i;
const unq = (v: string) => v.replace(/^"|"$/g, '');
const isQuoted = (v: string) => /^".*"$/.test(v);

export default function PalSettingsTab({server, onChanged}: TabProps) {
    const [opts, setOpts] = useState<backend.KV[] | null>(null);
    const [edits, setEdits] = useState<Record<string, string>>({});
    const [filter, setFilter] = useState('');

    const load = () => Host.GetPalSettings(server.id).then(o => { setOpts(o ?? []); setEdits({}); });
    useEffect(() => { load(); }, [server.id]);
    const shown = useMemo(() => (opts ?? []).filter(o => o.key.toLowerCase().includes(filter.toLowerCase())), [opts, filter]);
    if (!opts) return null;

    const value = (o: backend.KV) => edits[o.key] ?? unq(o.value);
    const save = () => {
        const out = opts.map(o => {
            if (!(o.key in edits)) return o;
            const v = edits[o.key];
            return new backend.KV({key: o.key, hasValue: true, value: isQuoted(o.value) ? `"${v.replace(/"/g, '')}"` : v});
        });
        run(Host.SetPalSettings(server.id, out), 'PalWorldSettings.ini saved').then(() => { load(); onChanged(); });
    };

    return (
        <Stack spacing={2}>
            <Typography variant="body2" sx={{color: 'text.secondary'}}>
                Edits OptionSettings in PalWorldSettings.ini. Values not set in the file show the defaults from DefaultPalWorldSettings.ini. Restart the server to apply.
            </Typography>
            <TextField placeholder="Filter settings" value={filter} onChange={e => setFilter(e.target.value)} sx={{maxWidth: 360}}/>
            <Paper sx={{p: 2}}>
                {!opts.length && <Typography variant="body2" sx={{color: 'text.secondary'}}>No settings found. Install the server first to load the defaults.</Typography>}
                <Box sx={{display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(260px, 1fr))', gap: 2}}>
                    {shown.map(o => /^(true|false)$/i.test(unq(o.value)) ? (
                        <TextField key={o.key} select label={o.key} value={value(o).toLowerCase() === 'true' ? 'True' : 'False'}
                                   onChange={e => setEdits({...edits, [o.key]: e.target.value})}>
                            <MenuItem value="True">True</MenuItem>
                            <MenuItem value="False">False</MenuItem>
                        </TextField>
                    ) : (
                        <TextField key={o.key} label={o.key} value={value(o)} type={SECRET.test(o.key) ? 'password' : 'text'}
                                   onChange={e => setEdits({...edits, [o.key]: e.target.value})}/>
                    ))}
                </Box>
            </Paper>
            <Stack direction="row" spacing={1} sx={{justifyContent: 'flex-end', position: 'sticky', bottom: 0, py: 1, bgcolor: 'background.paper'}}>
                <Button onClick={load} disabled={!Object.keys(edits).length}>Revert</Button>
                <Button variant="contained" onClick={save} disabled={!Object.keys(edits).length}>Save changes</Button>
            </Stack>
        </Stack>
    );
}
