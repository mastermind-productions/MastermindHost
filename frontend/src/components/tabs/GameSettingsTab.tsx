import {useEffect, useState} from 'react';
import {Box, Button, MenuItem, Paper, Stack, TextField, Typography} from '@mui/material';
import {backend, Host, notify, run} from '../../api';
import type {TabProps} from '../ServerPage';

type Def = {file: 'gusIni' | 'gameIni'; section: string; key: string; label: string; type: 'number' | 'bool' | 'text'};

const SS = '/Script/ShooterGame.ShooterGameMode';
const g = (key: string, label: string, type: Def['type'] = 'number', section = 'ServerSettings'): Def => ({file: 'gusIni', section, key, label, type});
const gm = (key: string, label: string, type: Def['type'] = 'number'): Def => ({file: 'gameIni', section: SS, key, label, type});

const GROUPS: [string, Def[]][] = [
    ['Server', [
        g('SessionName', 'Session name', 'text', 'SessionSettings'),
        g('ServerPassword', 'Join password', 'text'),
        g('ServerAdminPassword', 'Admin password', 'text'),
        g('MaxPlayers', 'Max players', 'number', '/Script/Engine.GameSession'),
        g('RCONEnabled', 'RCON enabled', 'bool'),
        g('RCONPort', 'RCON port'),
        g('ServerPVE', 'PvE', 'bool'),
        g('ServerHardcore', 'Hardcore', 'bool'),
        g('ServerCrosshair', 'Crosshair', 'bool'),
        g('ShowMapPlayerLocation', 'Show player on map', 'bool'),
        g('AllowThirdPersonPlayer', 'Third person', 'bool'),
        g('AlwaysNotifyPlayerJoined', 'Notify player joined', 'bool'),
        g('AlwaysNotifyPlayerLeft', 'Notify player left', 'bool'),
        g('KickIdlePlayersPeriod', 'Kick idle after (sec)'),
        g('AutoSavePeriodMinutes', 'Autosave period (min)'),
        g('ActiveMods', 'ActiveMods', 'text'),
    ]],
    ['Rates', [
        g('DifficultyOffset', 'Difficulty offset'),
        g('OverrideOfficialDifficulty', 'Override official difficulty'),
        g('XPMultiplier', 'XP multiplier'),
        g('TamingSpeedMultiplier', 'Taming speed'),
        g('HarvestAmountMultiplier', 'Harvest amount'),
        g('HarvestHealthMultiplier', 'Harvest health'),
        g('ResourcesRespawnPeriodMultiplier', 'Resource respawn period'),
        g('DayCycleSpeedScale', 'Day cycle speed'),
        g('DayTimeSpeedScale', 'Day time speed'),
        g('NightTimeSpeedScale', 'Night time speed'),
        g('PlayerCharacterFoodDrainMultiplier', 'Player food drain'),
        g('PlayerCharacterWaterDrainMultiplier', 'Player water drain'),
        g('PlayerCharacterStaminaDrainMultiplier', 'Player stamina drain'),
        g('DinoCharacterFoodDrainMultiplier', 'Dino food drain'),
        g('DinoCountMultiplier', 'Dino count'),
        g('StructureDamageMultiplier', 'Structure damage'),
        g('StructureResistanceMultiplier', 'Structure resistance'),
        g('ItemStackSizeMultiplier', 'Item stack size'),
    ]],
    ['Breeding (Game.ini)', [
        gm('MatingIntervalMultiplier', 'Mating interval'),
        gm('EggHatchSpeedMultiplier', 'Egg hatch speed'),
        gm('BabyMatureSpeedMultiplier', 'Baby mature speed'),
        gm('BabyFoodConsumptionSpeedMultiplier', 'Baby food consumption'),
        gm('BabyCuddleIntervalMultiplier', 'Cuddle interval'),
        gm('BabyImprintAmountMultiplier', 'Imprint amount'),
        gm('LayEggIntervalMultiplier', 'Lay egg interval'),
        gm('CropGrowthSpeedMultiplier', 'Crop growth speed'),
        gm('bAllowUnlimitedRespecs', 'Unlimited respecs', 'bool'),
        gm('bDisableStructurePlacementCollision', 'No placement collision', 'bool'),
        gm('bAllowFlyerSpeedLeveling', 'Flyer speed leveling', 'bool'),
    ]],
];

type Inis = Record<'gusIni' | 'gameIni', Record<string, Record<string, string>>>;
const id = (d: Def) => `${d.file}|${d.section}|${d.key}`;

export default function GameSettingsTab({server, onChanged}: TabProps) {
    const [inis, setInis] = useState<Inis | null>(null);
    const [edits, setEdits] = useState<Record<string, string>>({});

    const load = async () => {
        const [gus, game] = await Promise.all([Host.GetIni(server.id, 'gusIni'), Host.GetIni(server.id, 'gameIni')]);
        setInis({gusIni: gus ?? {}, gameIni: game ?? {}});
        setEdits({});
    };
    useEffect(() => { load(); }, [server.id]);
    if (!inis) return null;

    const value = (d: Def) => edits[id(d)] ?? inis[d.file][d.section]?.[d.key] ?? '';
    const save = async () => {
        for (const file of ['gusIni', 'gameIni'] as const) {
            const changes = GROUPS.flatMap(([, defs]) => defs)
                .filter(d => d.file === file && id(d) in edits)
                .map(d => new backend.IniChange({section: d.section, key: d.key, value: edits[id(d)], remove: edits[id(d)].trim() === ''}));
            if (!changes.length) continue;
            try {
                await Host.ApplyIni(server.id, file, changes);
            } catch (e: any) {
                notify(String(e?.message ?? e), 'error');
                return;
            }
        }
        await load();
        onChanged();
        notify('Settings saved', 'success');
    };

    return (
        <Stack spacing={2}>
            <Typography variant="body2" sx={{color: 'text.secondary'}}>
                Edits GameUserSettings.ini / Game.ini in place (unknown keys are preserved). Leave a field empty to remove the key. Command line options override ini values.
            </Typography>
            {GROUPS.map(([title, defs]) => (
                <Paper key={title} sx={{p: 2}}>
                    <Typography variant="subtitle1" sx={{mb: 2, fontWeight: 600}}>{title}</Typography>
                    <Box sx={{display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(220px, 1fr))', gap: 2}}>
                        {defs.map(d => d.type === 'bool' ? (
                            <TextField key={id(d)} select label={d.label} value={value(d).toLowerCase()} onChange={e => setEdits({...edits, [id(d)]: e.target.value === '' ? '' : e.target.value === 'true' ? 'True' : 'False'})}>
                                <MenuItem value="">(default)</MenuItem>
                                <MenuItem value="true">True</MenuItem>
                                <MenuItem value="false">False</MenuItem>
                            </TextField>
                        ) : (
                            <TextField key={id(d)} label={d.label} value={value(d)} type={d.type === 'number' ? 'number' : 'text'}
                                       onChange={e => setEdits({...edits, [id(d)]: e.target.value})}/>
                        ))}
                    </Box>
                </Paper>
            ))}
            <Stack direction="row" spacing={1} sx={{justifyContent: 'flex-end', position: 'sticky', bottom: 0, py: 1, bgcolor: 'background.paper'}}>
                <Button onClick={load} disabled={!Object.keys(edits).length}>Revert</Button>
                <Button variant="contained" onClick={save} disabled={!Object.keys(edits).length}>Save changes</Button>
            </Stack>
        </Stack>
    );
}
