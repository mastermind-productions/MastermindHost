import {useEffect, useState} from 'react';
import {Autocomplete, Box, Button, Checkbox, FormControlLabel, IconButton, Paper, Stack, TextField, Typography} from '@mui/material';
import DeleteIcon from '@mui/icons-material/Delete';
import {backend, Host, run} from '../../api';
import type {TabProps} from '../ServerPage';

const ARK_OPTS: [string, string, string?][] = [
    ['SessionName', 'Session name'],
    ['Port', 'Game port (UDP)', 'number'],
    ['RCONPort', 'RCON port', 'number'],
    ['ServerAdminPassword', 'Admin password'],
    ['ServerPassword', 'Join password'],
    ['MaxPlayers', 'Max players (option)', 'number'],
];

const PAL_VALUE_FLAGS: [string, string][] = [
    ['port', 'Game port (UDP)'],
    ['players', 'Max players'],
    ['publicip', 'Public IP'],
    ['publicport', 'Public port'],
    ['NumberOfWorkerThreadsServer', 'Worker threads'],
];

const PAL_BOOL_FLAGS: [string, string][] = [
    ['useperfthreads', 'Use perf threads'],
    ['NoAsyncLoadingThread', 'No async loading thread'],
    ['UseMultithreadForDS', 'Multithread for DS'],
    ['publiclobby', 'Community server (public lobby)'],
];

const ARK_VALUE_FLAGS: [string, string][] = [
    ['WinLiveMaxPlayers', 'Max players'],
    ['clusterid', 'Cluster ID'],
    ['ClusterDirOverride', 'Cluster dir override'],
    ['ServerPlatform', 'Server platform (e.g. ALL)'],
    ['culture', 'Culture (e.g. en)'],
];

const ARK_BOOL_FLAGS: [string, string][] = [
    ['servergamelog', 'Server game log'],
    ['NoBattlEye', 'Disable BattlEye'],
    ['exclusivejoin', 'Exclusive join (whitelist)'],
    ['ForceAllowCaveFlyers', 'Allow cave flyers'],
    ['NoTransferFromFiltering', 'No transfer filtering'],
    ['ForceRespawnDinos', 'Force respawn dinos'],
    ['UseDynamicConfig', 'Use dynamic config'],
    ['crossplay', 'Crossplay (legacy)'],
];

type Cmd = {map: string; options: backend.KV[]; flags: backend.KV[]};
const idx = (l: backend.KV[], k: string) => l.findIndex(x => x.key.toLowerCase() === k.toLowerCase());

export default function CommandLineTab({server, onChanged}: TabProps) {
    const [cmd, setCmd] = useState<Cmd | null>(null);
    const [raw, setRaw] = useState('');
    const [maps, setMaps] = useState<string[]>([]);
    const pal = server.game === 'palworld';
    const OPTS = pal ? [] : ARK_OPTS;
    const VALUE_FLAGS = pal ? PAL_VALUE_FLAGS : ARK_VALUE_FLAGS;
    const BOOL_FLAGS = pal ? PAL_BOOL_FLAGS : ARK_BOOL_FLAGS;
    useEffect(() => { Host.GetGames().then(gs => setMaps(gs.find(g => g.id === server.game)?.maps ?? [])); }, [server.game]);

    const load = () => {
        Host.GetParsedCmd(server.id).then(c => setCmd({map: c.map, options: c.options ?? [], flags: c.flags ?? []}));
        Host.ReadConfig(server.id, 'cmd').then(setRaw);
    };
    useEffect(load, [server.id]);
    if (!cmd) return null;

    const getOpt = (k: string) => { const i = idx(cmd.options, k); return i >= 0 ? cmd.options[i].value.replace(/^"|"$/g, '') : ''; };
    const setOpt = (k: string, v: string) => {
        const options = cmd.options.filter(o => o.key.toLowerCase() !== k.toLowerCase());
        const i = idx(cmd.options, k);
        if (v !== '') {
            const val = /\s/.test(v) ? `"${v}"` : v;
            const kv = new backend.KV({key: k, value: val, hasValue: true});
            i >= 0 ? options.splice(i, 0, kv) : options.push(kv);
        }
        setCmd({...cmd, options});
    };
    const getFlag = (k: string) => { const i = idx(cmd.flags, k); return i >= 0 ? cmd.flags[i] : null; };
    const setFlag = (k: string, v: string | null, hasValue: boolean) => {
        const flags = cmd.flags.filter(f => f.key.toLowerCase() !== k.toLowerCase());
        const i = idx(cmd.flags, k);
        if (v !== null && !(hasValue && v === '')) {
            const kv = new backend.KV({key: k, value: v, hasValue});
            i >= 0 ? flags.splice(i, 0, kv) : flags.push(kv);
        }
        setCmd({...cmd, flags});
    };
    const known = new Set([...OPTS.map(o => o[0]), 'RCONEnabled'].map(s => s.toLowerCase()));
    const knownFlags = new Set([...VALUE_FLAGS, ...BOOL_FLAGS].map(f => f[0].toLowerCase()).concat(['mods', 'passivemods']));
    const save = () => run(Host.SetParsedCmd(server.id, new backend.ArkCmd(cmd)), 'Command line saved').then(() => { load(); onChanged(); });

    return (
        <Stack spacing={2}>
            <Paper sx={{p: 2}}>
                <Typography variant="subtitle1" sx={{mb: 2, fontWeight: 600}}>Launch options</Typography>
                <Box sx={{display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(240px, 1fr))', gap: 2}}>
                    {!pal && <Autocomplete freeSolo options={maps} value={cmd.map} onInputChange={(_, v) => setCmd({...cmd, map: v})}
                                  renderInput={p => <TextField {...p} label="Map" size="small" helperText={server.game === 'arksa' && cmd.map === 'BobsMissions_WP' ? 'Club ARK requires CurseForge mod ID 1005639.' : undefined}/>}/>}
                    {OPTS.map(([k, label, type]) => (
                        <TextField key={k} label={label} type={type ?? 'text'} value={getOpt(k)} onChange={e => setOpt(k, e.target.value)}/>
                    ))}
                    {VALUE_FLAGS.map(([k, label]) => (
                        <TextField key={k} label={label} value={getFlag(k)?.value ?? ''} onChange={e => setFlag(k, e.target.value, true)}/>
                    ))}
                </Box>
                <Box sx={{display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(240px, 1fr))', mt: 1}}>
                    {!pal && <FormControlLabel label="RCON enabled" control={<Checkbox checked={getOpt('RCONEnabled').toLowerCase() === 'true'}
                                                                              onChange={e => setOpt('RCONEnabled', e.target.checked ? 'True' : 'False')}/>}/>}
                    {BOOL_FLAGS.map(([k, label]) => (
                        <FormControlLabel key={k} label={label} control={<Checkbox checked={!!getFlag(k)} onChange={e => setFlag(k, e.target.checked ? '' : null, false)}/>}/>
                    ))}
                </Box>
                <Typography variant="subtitle2" sx={{mt: 2, mb: 1}}>Other options / flags</Typography>
                <Stack spacing={1}>
                    {cmd.options.map((o, i) => known.has(o.key.toLowerCase()) ? null : (
                        <Row key={'o' + i} prefix="?" kv={o} onChange={kv => { const options = [...cmd.options]; options[i] = kv; setCmd({...cmd, options}); }}
                             onDelete={() => setCmd({...cmd, options: cmd.options.filter((_, j) => j !== i)})}/>
                    ))}
                    {cmd.flags.map((f, i) => knownFlags.has(f.key.toLowerCase()) ? null : (
                        <Row key={'f' + i} prefix="-" kv={f} onChange={kv => { const flags = [...cmd.flags]; flags[i] = kv; setCmd({...cmd, flags}); }}
                             onDelete={() => setCmd({...cmd, flags: cmd.flags.filter((_, j) => j !== i)})}/>
                    ))}
                    <Stack direction="row" spacing={1}>
                        <Button onClick={() => setCmd({...cmd, options: [...cmd.options, new backend.KV({key: 'NewOption', value: '', hasValue: true})]})}>+ ?Option</Button>
                        <Button onClick={() => setCmd({...cmd, flags: [...cmd.flags, new backend.KV({key: 'NewFlag', value: '', hasValue: false})]})}>+ -Flag</Button>
                        <Box sx={{flex: 1}}/>
                        <Button onClick={load}>Revert</Button>
                        <Button variant="contained" onClick={save}>Save</Button>
                    </Stack>
                </Stack>
            </Paper>
            <Paper sx={{p: 2}}>
                <Typography variant="subtitle1" sx={{mb: 1, fontWeight: 600}}>Raw command line</Typography>
                <TextField fullWidth multiline minRows={3} value={raw} onChange={e => setRaw(e.target.value)}
                           sx={{'& textarea': {fontFamily: 'Consolas, monospace', fontSize: 13}}}/>
                <Stack direction="row" sx={{justifyContent: 'flex-end', mt: 1}}>
                    <Button variant="contained" onClick={() => run(Host.SetCmd(server.id, raw), 'Command line saved').then(() => { load(); onChanged(); })}>Save raw</Button>
                </Stack>
            </Paper>
        </Stack>
    );
}

function Row({prefix, kv, onChange, onDelete}: {prefix: string; kv: backend.KV; onChange: (kv: backend.KV) => void; onDelete: () => void}) {
    return (
        <Stack direction="row" spacing={1} sx={{alignItems: 'center'}}>
            <Typography sx={{width: 12, color: 'primary.main', fontFamily: 'monospace'}}>{prefix}</Typography>
            <TextField label="Key" value={kv.key} onChange={e => onChange(new backend.KV({...kv, key: e.target.value}))} sx={{width: 260}}/>
            <TextField label="Value" value={kv.value} disabled={!kv.hasValue} onChange={e => onChange(new backend.KV({...kv, value: e.target.value}))} sx={{flex: 1}}/>
            <FormControlLabel label="Has value" control={<Checkbox checked={kv.hasValue} onChange={e => onChange(new backend.KV({...kv, hasValue: e.target.checked}))}/>}/>
            <IconButton onClick={onDelete}><DeleteIcon/></IconButton>
        </Stack>
    );
}
