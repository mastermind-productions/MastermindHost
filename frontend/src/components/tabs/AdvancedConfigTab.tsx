import {memo, useCallback, useEffect, useMemo, useRef, useState} from 'react';
import {Box, Button, IconButton, InputAdornment, MenuItem, Paper, Stack, TextField, Typography} from '@mui/material';
import ClearIcon from '@mui/icons-material/Clear';
import {backend, Host, notify, run} from '../../api';
import type {TabProps} from '../ServerPage';

const SECRET = /password/i;
const PAGE = 120;

const eid = (e: backend.AdvancedIniEntry) => `${e.file}|${e.section}|${e.key}`;

type RowProps = {e: backend.AdvancedIniEntry; value: string; onEdit: (e: backend.AdvancedIniEntry, v: string) => void};

const EntryRow = memo(function EntryRow({e, value, onEdit}: RowProps) {
    const v = value;
    const sub = e.desc
        ? (e.set ? e.desc : `${e.desc} - adds to ${e.fileName}`)
        : (e.set ? `In ${e.fileName} [${e.section}]` : `Not set - adds to ${e.fileName} [${e.section}]`);
    if (e.kind === 'bool') {
        const bv = v === '' ? '' : /^(true|1)$/i.test(v.trim()) ? 'True' : 'False';
        return (
            <TextField select label={e.key} helperText={sub} size="small" fullWidth value={bv}
                       onChange={ev => onEdit(e, ev.target.value)}>
                <MenuItem value="">{e.set ? '(remove)' : '(not set)'}</MenuItem>
                <MenuItem value="True">True</MenuItem>
                <MenuItem value="False">False</MenuItem>
            </TextField>
        );
    }
    const multiline = e.kind === 'group' || v.length > 80 || v.includes('\n');
    return (
        <TextField label={e.key} helperText={sub} size="small" fullWidth multiline={multiline}
                   minRows={multiline ? 3 : undefined} spellCheck={false} value={v}
                   type={SECRET.test(e.key) ? 'password' : undefined}
                   placeholder={e.set ? '' : '(not set)'}
                   onChange={ev => onEdit(e, ev.target.value)}
                   sx={{'& .MuiFormHelperText-root': {whiteSpace: 'normal'}}}
                   slotProps={{
                       input: {
                           endAdornment: e.set && v !== '' ? (
                               <InputAdornment position="end" sx={{alignSelf: 'flex-start'}}>
                                   <IconButton size="small" title="Remove key" onClick={() => onEdit(e, '')}>
                                       <ClearIcon fontSize="small"/>
                                   </IconButton>
                               </InputAdornment>
                           ) : undefined,
                       },
                   }}/>
    );
});

export default function AdvancedConfigTab({server, onChanged}: TabProps) {
    const [entries, setEntries] = useState<backend.AdvancedIniEntry[] | null>(null);
    const [edits, setEdits] = useState<Record<string, string>>({});
    const [filter, setFilter] = useState('');
    const [query, setQuery] = useState('');

    const load = () => Host.AdvancedIniList(server.id).then(e => { setEntries(e ?? []); setEdits({}); });
    useEffect(() => { load(); }, [server.id]);

    useEffect(() => {
        const t = setTimeout(() => setQuery(filter), 150);
        return () => clearTimeout(t);
    }, [filter]);

    const cur = (e: backend.AdvancedIniEntry) => edits[eid(e)] ?? (e.set ? e.value : '');
    const setEdit = useCallback((e: backend.AdvancedIniEntry, v: string) => {
        setEdits(prev => ({...prev, [`${e.file}|${e.section}|${e.key}`]: v}));
    }, []);

    const shown = useMemo(() => {
        const tokens = query.toLowerCase().split(/\s+/).filter(Boolean);
        return (entries ?? []).filter(e => {
            const hay = `${e.key} ${e.desc} ${e.group} ${e.section}`.toLowerCase();
            return tokens.every(t => hay.includes(t));
        });
    }, [entries, query]);

    const [limit, setLimit] = useState(PAGE);
    const sentinelRef = useRef<HTMLDivElement | null>(null);
    useEffect(() => { setLimit(PAGE); }, [query, entries]);
    useEffect(() => {
        const el = sentinelRef.current;
        if (!el || !entries || shown.length <= limit) return;
        const io = new IntersectionObserver(
            es => { if (es.some(x => x.isIntersecting)) setLimit(l => l + PAGE); },
            {rootMargin: '1200px'});
        io.observe(el);
        return () => io.disconnect();
    }, [entries, shown, limit]);
    if (!entries) return null;

    const visible = shown.slice(0, limit);
    const groups: {label: string; items: backend.AdvancedIniEntry[]}[] = [];
    for (const e of visible) {
        const label = e.group || e.section;
        const last = groups[groups.length - 1];
        if (last?.label === label) last.items.push(e);
        else groups.push({label, items: [e]});
    }

    const dirtyCount = Object.keys(edits).length;
    const save = async () => {
        const changes = entries.filter(e => eid(e) in edits).map(e => {
            const v = edits[eid(e)];
            return new backend.AdvancedIniChange({file: e.file, section: e.section, key: e.key, value: v, remove: v.trim() === '' && e.set});
        }).filter(c => c.remove || c.value.trim() !== '');
        try {
            await Host.AdvancedIniApply(server.id, changes);
            notify('Settings saved', 'success');
            load();
            onChanged();
        } catch (e: any) {
            notify(String(e?.message ?? e), 'error');
        }
    };

    return (
        <Stack spacing={1.5}>
            <Typography variant="body2" sx={{color: 'text.secondary'}}>
                Every key in the server's .ini files (GameUserSettings.ini / Game.ini), searchable and editable. Rows marked
                "(not set)" come from the built-in catalog - fill in a value to add them; clearing a value removes the key.
                Stop the server before saving: a running server overwrites these files on shutdown.
            </Typography>
            <Stack direction="row" spacing={2} sx={{alignItems: 'center', position: 'sticky', top: -8, zIndex: 2, bgcolor: 'background.paper', py: 1, flexWrap: 'wrap', rowGap: 1}}>
                <TextField placeholder='Search settings... (e.g. "dino weight", "imprint", "pvp")' value={filter} size="small"
                           onChange={e => setFilter(e.target.value)} sx={{flex: 1, minWidth: 280}}/>
                <Typography variant="body2" sx={{color: 'text.secondary'}}>{shown.length} / {entries.length}</Typography>
                <Button onClick={() => setEdits({})} disabled={!dirtyCount}>Revert</Button>
                <Button variant="contained" onClick={save} disabled={!dirtyCount}>Save changes{dirtyCount ? ` (${dirtyCount})` : ''}</Button>
            </Stack>
            {groups.map(gr => (
                <Paper key={gr.label} sx={{p: 2}}>
                    <Typography variant="subtitle1" sx={{mb: 1.5, fontWeight: 600}}>{gr.label}</Typography>
                    <Box sx={{display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(360px, 1fr))', gap: 2}}>
                        {gr.items.map(e => <EntryRow key={eid(e)} e={e} value={cur(e)} onEdit={setEdit}/>)}
                    </Box>
                </Paper>
            ))}
            <div ref={sentinelRef} style={{height: 8}}/>
        </Stack>
    );
}
