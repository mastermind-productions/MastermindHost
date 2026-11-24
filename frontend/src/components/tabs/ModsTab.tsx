import {useEffect, useState} from 'react';
import {Button, CircularProgress, IconButton, Link, Paper, Stack, Table, TableBody, TableCell, TableHead, TableRow, TextField, Typography} from '@mui/material';
import DeleteIcon from '@mui/icons-material/Delete';
import ArrowUpwardIcon from '@mui/icons-material/ArrowUpward';
import ArrowDownwardIcon from '@mui/icons-material/ArrowDownward';
import {backend, Host, run} from '../../api';
import {BrowserOpenURL} from '../../../wailsjs/runtime/runtime';
import type {TabProps} from '../ServerPage';

export default function ModsTab({server, onChanged}: TabProps) {
    const [mods, setMods] = useState<backend.ModInfo[]>([]);
    const [newId, setNewId] = useState('');
    const [newName, setNewName] = useState('');
    const [searchName, setSearchName] = useState('');
    const [searchResults, setSearchResults] = useState<backend.CurseForgeMod[]>([]);
    const [searching, setSearching] = useState(false);
    const [searchError, setSearchError] = useState('');
    const [dirty, setDirty] = useState(false);

    const load = () => Host.GetMods(server.id).then(m => {
        setMods(m); setDirty(false);
        if (server.game === 'arksa' && m.some(x => !x.name)) {
            Host.ResolveModNames(server.id).then(n => { if (n > 0) Host.GetMods(server.id).then(r => { setMods(r); setDirty(false); }); }).catch(() => {});
        }
    });
    useEffect(() => { load(); }, [server.id]);

    const update = (m: backend.ModInfo[]) => { setMods(m); setDirty(true); };
    const move = (i: number, d: number) => {
        const m = [...mods];
        [m[i], m[i + d]] = [m[i + d], m[i]];
        update(m);
    };
    const add = () => {
        const ids = newId.split(/[\s,]+/).map(s => s.trim()).filter(s => /^\d+$/.test(s) && !mods.some(m => m.id === s));
        if (!ids.length) return;
        update([...mods, ...ids.map(id => new backend.ModInfo({id, name: ids.length === 1 ? newName : ''}))]);
        setNewId(''); setNewName('');
    };
    const searchByName = async () => {
        setSearching(true);
        setSearchError('');
        try {
            setSearchResults(await Host.SearchCurseForgeMods(server.id, searchName));
        } catch (e: any) {
            setSearchResults([]);
            setSearchError(String(e?.message ?? e));
        } finally {
            setSearching(false);
        }
    };
    const addSearchResult = (result: backend.CurseForgeMod) => {
        if (mods.some(m => m.id === result.id)) return;
        update([...mods, new backend.ModInfo({id: result.id, name: result.name})]);
    };

    return (
        <Stack spacing={2}>
            <Typography variant="body2" sx={{color: 'text.secondary'}}>
                Search CurseForge live by mod name, or add project IDs directly. Search results are fetched from CurseForge each time; no bundled catalogue can go stale. Mods download automatically when the server starts. Load order is top to bottom.
            </Typography>
            <Stack direction="row" sx={{justifyContent: 'flex-end'}}>
                <Button variant="outlined" onClick={() => BrowserOpenURL('https://www.curseforge.com/ark-survival-ascended/search')}>Browse CurseForge</Button>
            </Stack>
            <Paper sx={{p: 2}}>
                <Stack spacing={1.5}>
                    <Stack direction="row" spacing={1}>
                        <TextField label="Find mod by name" value={searchName} onChange={e => setSearchName(e.target.value)} onKeyDown={e => e.key === 'Enter' && searchByName()} sx={{flex: 1}}/>
                        <Button variant="contained" disabled={searching || searchName.trim().length < 2} onClick={searchByName}>
                            {searching ? <CircularProgress size={20}/> : 'Search'}
                        </Button>
                    </Stack>
                    {searchError && <Typography color="error" variant="body2">{searchError}</Typography>}
                    {searchResults.length > 0 && <Stack spacing={0.5}>
                        <Typography variant="subtitle2">Live CurseForge results</Typography>
                        {searchResults.map(result => {
                            const alreadyAdded = mods.some(m => m.id === result.id);
                            return <Stack key={result.id} direction="row" spacing={1} sx={{alignItems: 'center', borderTop: 1, borderColor: 'divider', pt: 0.75}}>
                                <Typography sx={{flex: 1}}>{result.name}</Typography>
                                <Typography variant="body2" color="text.secondary">ID {result.id}</Typography>
                                <Link component="button" onClick={() => BrowserOpenURL(result.url)}>View</Link>
                                <Button size="small" disabled={alreadyAdded} onClick={() => addSearchResult(result)}>{alreadyAdded ? 'Added' : 'Add'}</Button>
                            </Stack>;
                        })}
                    </Stack>}
                    {!searching && searchName.trim().length >= 2 && searchResults.length === 0 && !searchError && <Typography variant="body2" color="text.secondary">No matching mods found.</Typography>}
                </Stack>
            </Paper>
            <Paper sx={{p: 2}}>
                <Stack direction="row" spacing={1}>
                    <TextField label="Mod ID(s), comma separated" value={newId} onChange={e => setNewId(e.target.value)} onKeyDown={e => e.key === 'Enter' && add()} sx={{width: 300}}/>
                    <TextField label="Name (optional)" value={newName} onChange={e => setNewName(e.target.value)} onKeyDown={e => e.key === 'Enter' && add()} sx={{flex: 1}}/>
                    <Button variant="outlined" onClick={add}>Add by ID</Button>
                </Stack>
            </Paper>
            <Paper>
                <Table size="small">
                    <TableHead>
                        <TableRow><TableCell>#</TableCell><TableCell>Mod ID</TableCell><TableCell>Name</TableCell><TableCell align="right">Actions</TableCell></TableRow>
                    </TableHead>
                    <TableBody>
                        {mods.map((m, i) => (
                            <TableRow key={m.id}>
                                <TableCell>{i + 1}</TableCell>
                                <TableCell>
                                    <Link component="button" onClick={() => BrowserOpenURL(`https://www.curseforge.com/ark-survival-ascended/search?search=${m.id}`)}>{m.id}</Link>
                                </TableCell>
                                <TableCell>
                                    <TextField variant="standard" value={m.name} placeholder="(unnamed)" fullWidth
                                               onChange={e => update(mods.map((x, j) => j === i ? new backend.ModInfo({...x, name: e.target.value}) : x))}/>
                                </TableCell>
                                <TableCell align="right" sx={{whiteSpace: 'nowrap'}}>
                                    <IconButton size="small" disabled={i === 0} onClick={() => move(i, -1)}><ArrowUpwardIcon fontSize="small"/></IconButton>
                                    <IconButton size="small" disabled={i === mods.length - 1} onClick={() => move(i, 1)}><ArrowDownwardIcon fontSize="small"/></IconButton>
                                    <IconButton size="small" onClick={() => update(mods.filter((_, j) => j !== i))}><DeleteIcon fontSize="small"/></IconButton>
                                </TableCell>
                            </TableRow>
                        ))}
                        {!mods.length && <TableRow><TableCell colSpan={4} sx={{color: 'text.secondary'}}>No mods installed.</TableCell></TableRow>}
                    </TableBody>
                </Table>
            </Paper>
            <Stack direction="row" spacing={1} sx={{justifyContent: 'flex-end'}}>
                <Button onClick={load} disabled={!dirty}>Revert</Button>
                <Button variant="contained" disabled={!dirty} onClick={() => run(Host.SetMods(server.id, mods), 'Mods saved').then(() => { load(); onChanged(); })}>Save mods</Button>
            </Stack>
        </Stack>
    );
}
