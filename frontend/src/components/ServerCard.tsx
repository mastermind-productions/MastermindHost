import {Box, Button, Card, CardContent, CardMedia, Chip, IconButton, LinearProgress, Stack, Tooltip, Typography} from '@mui/material';
import heroArt from '../assets/images/hero.jpg';
import rappelzArt from '../assets/images/rappelz.jpg';
import scionsArt from '../assets/images/scions.jpg';
import TuneIcon from '@mui/icons-material/Tune';
import ChevronLeftIcon from '@mui/icons-material/ChevronLeft';
import ChevronRightIcon from '@mui/icons-material/ChevronRight';
import {backend, fmtBytes, fmtUptime, statusColor} from '../api';
import Actions from './Actions';

const ART: Record<string, string> = {hero: heroArt, rappelz: rappelzArt, scions: scionsArt};

export type OrderProps = {
    onMove: (delta: -1 | 1) => void;
    canPrev: boolean;
    canNext: boolean;
    dragging: boolean;
    onDragStart: () => void;
    onDragEnd: () => void;
    onDropHere: () => void;
};

export default function ServerCard({server: s, onConfigure, onChanged, order}: {server: backend.ServerInfo; onConfigure: () => void; onChanged: () => void; order?: OrderProps}) {
    const header = ART[s.game] || s.header;
    const working = s.status === 'installing' || s.status === 'updating';
    return (
        <Card sx={{display: 'flex', flexDirection: 'column', opacity: order?.dragging ? 0.4 : 1}}
              draggable={!!order} onDragStart={order?.onDragStart} onDragEnd={order?.onDragEnd}
              onDragOver={order ? (e => e.preventDefault()) : undefined} onDrop={order ? (e => { e.preventDefault(); order.onDropHere(); }) : undefined}>
            <Box sx={{position: 'relative'}}>
                {header ? <CardMedia component="img" height="110" image={header} alt={s.gameName} sx={{filter: 'brightness(0.45) saturate(0.8)'}}/>
                    : <Box sx={{height: 110, background: 'linear-gradient(135deg, #0b1f13 0%, #14532d 60%, #050806 100%)'}}/>}
                <Box sx={{position: 'absolute', inset: 0, p: 1.5, display: 'flex', flexDirection: 'column', justifyContent: 'space-between'}}>
                    <Stack direction="row" spacing={1} sx={{justifyContent: 'space-between'}}>
                        <Chip size="small" label={s.gameName}/>
                        <Chip size="small" color={statusColor(s.status)} label={s.status.toUpperCase()}/>
                    </Stack>
                    <Typography variant="h6" sx={{fontWeight: 700, textShadow: '0 1px 3px #000'}} noWrap title={s.name}>{s.name}</Typography>
                </Box>
            </Box>
            {working && <LinearProgress variant={s.progress > 0 ? 'determinate' : 'indeterminate'} value={s.progress}/>}
            <CardContent sx={{flex: 1, display: 'flex', flexDirection: 'column', gap: 1.5}}>
                <Box sx={{display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', rowGap: 1, columnGap: 2}}>
                    {s.script ? (<>
                        <Field label="Components" value={`${s.compUp} / ${s.compTotal}`}/>
                        <Field label="Game port" value={s.port ? String(s.port) : '-'}/>
                        <Field label="Uptime" value={fmtUptime(s.startTime)}/>
                        <Field label="CPU" value={s.startTime ? `${s.cpu.toFixed(1)}%` : '-'}/>
                        <Field label="RAM" value={s.startTime ? fmtBytes(s.ram) : '-'}/>
                    </>) : (<>
                    <Field label="Players" value={`${s.playerCount} / ${s.maxPlayers}`}/>
                    <Field label="Map" value={s.map || '-'}/>
                    <Field label="Port" value={String(s.port)}/>
                    <Field label="CPU" value={s.startTime ? `${s.cpu.toFixed(1)}%` : '-'}/>
                    <Field label="RAM" value={s.startTime ? fmtBytes(s.ram) : '-'}/>
                    <Field label="Uptime" value={fmtUptime(s.startTime)}/>
                    <Field label="Build" value={s.buildId || '-'}/>
                    <Field label="Mods" value={String(s.modCount)}/>
                    <Field label="Auto-restart" value={s.forceRunning ? 'On' : 'Off'}/>
                    </>)}
                </Box>
                <Box sx={{flex: 1}}/>
                <Stack direction="row" sx={{justifyContent: 'space-between', alignItems: 'flex-end', gap: 1}}>
                    <Actions server={s} onChanged={onChanged}/>
                    <Stack direction="row" sx={{alignItems: 'center'}}>
                        {order && <Tooltip title="Move earlier"><span><IconButton size="small" disabled={!order.canPrev} onClick={() => order.onMove(-1)}><ChevronLeftIcon/></IconButton></span></Tooltip>}
                        {order && <Tooltip title="Move later"><span><IconButton size="small" disabled={!order.canNext} onClick={() => order.onMove(1)}><ChevronRightIcon/></IconButton></span></Tooltip>}
                        <Button variant="text" startIcon={<TuneIcon/>} onClick={onConfigure}>Configure</Button>
                    </Stack>
                </Stack>
            </CardContent>
        </Card>
    );
}

export function Field({label, value}: {label: string; value: string}) {
    return (
        <Box sx={{minWidth: 0}}>
            <Typography variant="caption" sx={{color: 'text.secondary'}}>{label}</Typography>
            <Typography variant="body2" noWrap title={value} sx={{fontWeight: 600}}>{value}</Typography>
        </Box>
    );
}
