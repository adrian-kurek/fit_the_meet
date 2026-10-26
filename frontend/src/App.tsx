import { FormEvent, PointerEvent as ReactPointerEvent, useEffect, useMemo, useRef, useState } from 'react';
import {
  ActionIcon,
  Anchor,
  Avatar,
  Badge,
  Box,
  Burger,
  Button,
  Checkbox,
  Divider,
  Drawer,
  Group,
  Menu,
  Modal,
  NumberInput,
  Paper,
  PasswordInput,
  SegmentedControl,
  Select,
  Stack,
  Text,
  Textarea,
  TextInput,
  ThemeIcon,
  Title,
  UnstyledButton,
} from '@mantine/core';
import { useDisclosure } from '@mantine/hooks';
import {
  ApiError,
  authApi,
  AvailabilityTimeBlock,
  eventApi,
  groupApi,
  LoginCredentials,
  pollApi,
  RegistrationDetails,
} from './api';

type IconName =
  | 'calendar'
  | 'check'
  | 'clock'
  | 'filter'
  | 'group'
  | 'heart'
  | 'home'
  | 'lock'
  | 'location'
  | 'mail'
  | 'plus'
  | 'search'
  | 'settings'
  | 'sparkle'
  | 'user'
  | 'video';

type MeetingStatus = 'needs-response' | 'collecting' | 'ready' | 'scheduled';

type TimeOption = {
  id: number;
  label: string;
  detail: string;
  votes: number;
};

type Meeting = {
  id: number;
  shareCode: string;
  title: string;
  description: string;
  group: string;
  location: string;
  organizer: string;
  dateLabel: string;
  duration: string;
  responses: number;
  participants: number;
  status: MeetingStatus;
  accent: string;
  softAccent: string;
  icon: IconName;
  attendees: string[];
  favorite?: boolean;
  isActive?: boolean;
  apiId?: number | null;
  kind?: 'event' | 'poll';
  groupId?: number;
  options: TimeOption[];
};

type GroupParticipant = {
  id: number;
  name: string;
  email: string;
  initials: string;
};

type WorkspaceGroup = {
  id: number;
  name: string;
  color: string;
  role: 'owner' | 'member';
  participants: GroupParticipant[];
};

const slotMinutes = 30;
const availabilityStartMinutes = 8 * 60;
const availabilityEndMinutes = 20 * 60;

const availabilityTimes = Array.from(
  { length: (availabilityEndMinutes - availabilityStartMinutes) / slotMinutes },
  (_, index) => availabilityStartMinutes + (index * slotMinutes),
);

function availabilityFromOption(option: TimeOption): AvailabilityTimeBlock {
  const dateMatch = option.label.match(/(\d{1,2})\s(Oct|Nov)/);
  const month = dateMatch?.[2] === 'Nov' ? '11' : '10';
  const day = dateMatch?.[1].padStart(2, '0') || '26';
  const [startTime = '09:00', endTime = '10:00'] = option.detail.split('–');
  return {
    date: `2026-${month}-${day}`,
    start_time: startTime,
    end_time: endTime,
  };
}

function timeToMinutes(time: string) {
  const [hours = 0, minutes = 0] = time.split(':').map(Number);
  return (hours * 60) + minutes;
}

function minutesToTime(minutes: number) {
  return `${String(Math.floor(minutes / 60)).padStart(2, '0')}:${String(minutes % 60).padStart(2, '0')}`;
}

function slotKey(date: string, minutes: number) {
  return `${date}|${minutes}`;
}

function addDays(date: string, amount: number) {
  const next = new Date(`${date}T12:00:00Z`);
  next.setUTCDate(next.getUTCDate() + amount);
  return next.toISOString().slice(0, 10);
}

function dateWindow(startDate: string) {
  return Array.from({ length: 5 }, (_, index) => addDays(startDate, index));
}

function slotsFromBlocks(blocks: AvailabilityTimeBlock[]) {
  const selected = new Set<string>();
  blocks.forEach((block) => {
    const start = Math.max(availabilityStartMinutes, timeToMinutes(block.start_time));
    const end = Math.min(availabilityEndMinutes, timeToMinutes(block.end_time));
    for (let minutes = start; minutes < end; minutes += slotMinutes) {
      selected.add(slotKey(block.date, minutes));
    }
  });
  return selected;
}

function blocksFromSlots(selected: Set<string>): AvailabilityTimeBlock[] {
  const slotsByDate = new Map<string, number[]>();
  selected.forEach((key) => {
    const [date, minutes] = key.split('|');
    if (!date || minutes === undefined) return;
    slotsByDate.set(date, [...(slotsByDate.get(date) || []), Number(minutes)]);
  });

  const blocks: AvailabilityTimeBlock[] = [];
  [...slotsByDate.entries()].sort(([dateA], [dateB]) => dateA.localeCompare(dateB)).forEach(([date, minutes]) => {
    const sorted = [...new Set(minutes)].sort((a, b) => a - b);
    let start = sorted[0];
    let previous = sorted[0];
    sorted.slice(1).forEach((current) => {
      if (start === undefined || previous === undefined) return;
      if (current !== previous + slotMinutes) {
        blocks.push({ date, start_time: minutesToTime(start), end_time: minutesToTime(previous + slotMinutes) });
        start = current;
      }
      previous = current;
    });
    if (start !== undefined && previous !== undefined) {
      blocks.push({ date, start_time: minutesToTime(start), end_time: minutesToTime(previous + slotMinutes) });
    }
  });
  return blocks;
}

function formatDay(date: string) {
  return new Intl.DateTimeFormat('en', { weekday: 'short', day: 'numeric', month: 'short', timeZone: 'UTC' })
    .format(new Date(`${date}T12:00:00Z`));
}

function formatSelectedDuration(slotCount: number) {
  const totalMinutes = slotCount * slotMinutes;
  const hours = Math.floor(totalMinutes / 60);
  const minutes = totalMinutes % 60;
  return `${hours ? `${hours}h` : ''}${hours && minutes ? ' ' : ''}${minutes ? `${minutes}m` : ''}`;
}

const initialMeetings: Meeting[] = [
  {
    id: 1,
    shareCode: 'EVENT-DEMO',
    title: 'Project planning meeting',
    description: 'Choose a time for the next project meeting.',
    group: 'Project group',
    location: 'Warsaw · Studio 2',
    organizer: 'Maya K.',
    dateLabel: 'Closes today',
    duration: '50 min',
    responses: 8,
    participants: 10,
    status: 'needs-response',
    accent: '#e9346f',
    softAccent: '#fff0f5',
    icon: 'sparkle',
    attendees: ['MK', 'AP', 'JT', 'LN'],
    favorite: true,
    groupId: 101,
    options: [
      { id: 11, label: 'Monday, 26 Oct', detail: '10:00–10:50', votes: 8 },
      { id: 12, label: 'Tuesday, 27 Oct', detail: '14:00–14:50', votes: 6 },
      { id: 13, label: 'Wednesday, 28 Oct', detail: '09:00–09:50', votes: 5 },
    ],
  },
  {
    id: 2,
    shareCode: 'POLL-DEMO',
    title: 'Interface review',
    description: 'Review the mobile layouts and availability screens.',
    group: 'UI review group',
    location: 'Google Meet',
    organizer: 'Alex R.',
    dateLabel: 'Tomorrow',
    duration: '40 min',
    responses: 6,
    participants: 6,
    status: 'ready',
    accent: '#6f45ed',
    softAccent: '#f3efff',
    icon: 'video',
    kind: 'poll',
    groupId: 102,
    attendees: ['AR', 'AK', 'DB', 'IM'],
    options: [
      { id: 21, label: 'Tuesday, 27 Oct', detail: '11:30–12:10', votes: 6 },
      { id: 22, label: 'Tuesday, 27 Oct', detail: '15:00–15:40', votes: 4 },
    ],
  },
  {
    id: 3,
    shareCode: 'EVENT-WEEKLY',
    title: 'Weekly group meeting',
    description: 'Choose a time for this week’s group meeting.',
    group: 'Study group',
    location: 'Remote',
    organizer: 'Ola P.',
    dateLabel: '3 days left',
    duration: '60 min',
    responses: 4,
    participants: 7,
    status: 'collecting',
    accent: '#ff681d',
    softAccent: '#fff2ea',
    icon: 'group',
    groupId: 104,
    attendees: ['OP', 'MW', 'KS', 'AR'],
    options: [
      { id: 31, label: 'Thursday, 29 Oct', detail: '09:00–10:00', votes: 4 },
      { id: 32, label: 'Thursday, 29 Oct', detail: '16:00–17:00', votes: 3 },
      { id: 33, label: 'Friday, 30 Oct', detail: '10:00–11:00', votes: 2 },
    ],
  },
  {
    id: 4,
    shareCode: 'EVENT-API',
    title: 'API review',
    description: 'Review the event and poll API integration.',
    group: 'Developers',
    location: 'Warsaw · Room 4B',
    organizer: 'Adrian K.',
    dateLabel: 'Scheduled · Wed',
    duration: '45 min',
    responses: 9,
    participants: 9,
    status: 'scheduled',
    accent: '#078fa4',
    softAccent: '#eaf9fb',
    icon: 'check',
    attendees: ['AK', 'MK', 'PS', 'EW'],
    favorite: true,
    groupId: 103,
    options: [
      { id: 41, label: 'Wednesday, 28 Oct', detail: '13:00–13:45', votes: 9 },
    ],
  },
  {
    id: 5,
    shareCode: 'EVENT-INTERVIEW',
    title: 'User interview review',
    description: 'Discuss notes from recent user interviews.',
    group: 'Interview group',
    location: 'Remote',
    organizer: 'Nina S.',
    dateLabel: '5 days left',
    duration: '75 min',
    responses: 3,
    participants: 8,
    status: 'needs-response',
    accent: '#168a56',
    softAccent: '#ecfbf4',
    icon: 'user',
    attendees: ['NS', 'AP', 'JO', 'EN'],
    options: [
      { id: 51, label: 'Monday, 2 Nov', detail: '10:00–11:15', votes: 3 },
      { id: 52, label: 'Monday, 2 Nov', detail: '14:00–15:15', votes: 3 },
    ],
  },
  {
    id: 6,
    shareCode: 'EVENT-DINNER',
    title: 'Group dinner',
    description: 'Vote for a time, then choose the place.',
    group: 'Everyone',
    location: 'Warsaw · TBD',
    organizer: 'Group organizer',
    dateLabel: '1 week left',
    duration: '2 hours',
    responses: 12,
    participants: 18,
    status: 'collecting',
    accent: '#ffb000',
    softAccent: '#fff8df',
    icon: 'calendar',
    attendees: ['PO', 'AR', 'MK', 'NS'],
    groupId: 105,
    options: [
      { id: 61, label: 'Friday, 6 Nov', detail: '18:00–20:00', votes: 12 },
      { id: 62, label: 'Saturday, 7 Nov', detail: '17:00–19:00', votes: 9 },
    ],
  },
];

const filterItems = [
  { id: 'all', label: 'All events', icon: 'home' as IconName, color: '#252525' },
  { id: 'needs-response', label: 'Needs your response', icon: 'sparkle' as IconName, color: '#e9346f' },
  { id: 'this-week', label: 'This week', icon: 'calendar' as IconName, color: '#6f45ed' },
  { id: 'ready', label: 'Ready to schedule', icon: 'check' as IconName, color: '#078fa4' },
  { id: 'favorites', label: 'Favorites', icon: 'heart' as IconName, color: '#ff681d' },
];

const initialGroups: WorkspaceGroup[] = [
  {
    id: 101,
    name: 'Project group',
    color: '#e9346f',
    role: 'member',
    participants: [
      { id: 11, name: 'Maya Kowalska', email: 'maya@fitmeet.dev', initials: 'MK' },
      { id: 12, name: 'Jan Tomasz', email: 'jan@fitmeet.dev', initials: 'JT' },
    ],
  },
  {
    id: 102,
    name: 'UI review group',
    color: '#6f45ed',
    role: 'member',
    participants: [
      { id: 21, name: 'Iga Maj', email: 'iga@fitmeet.dev', initials: 'IM' },
      { id: 22, name: 'Dawid Baran', email: 'dawid@fitmeet.dev', initials: 'DB' },
    ],
  },
  {
    id: 103,
    name: 'Developers',
    color: '#078fa4',
    role: 'owner',
    participants: [
      { id: 31, name: 'Marta Kania', email: 'marta@fitmeet.dev', initials: 'MK' },
      { id: 32, name: 'Piotr Sikora', email: 'piotr@fitmeet.dev', initials: 'PS' },
      { id: 33, name: 'Ewa Wrona', email: 'ewa@fitmeet.dev', initials: 'EW' },
      { id: 34, name: 'Leon Nowak', email: 'leon@fitmeet.dev', initials: 'LN' },
    ],
  },
  {
    id: 104,
    name: 'Study group',
    color: '#ff681d',
    role: 'member',
    participants: [
      { id: 41, name: 'Ola Pawlak', email: 'ola@fitmeet.dev', initials: 'OP' },
      { id: 42, name: 'Kamil Serafin', email: 'kamil@fitmeet.dev', initials: 'KS' },
    ],
  },
  {
    id: 105,
    name: 'Everyone',
    color: '#ffb000',
    role: 'owner',
    participants: [
      { id: 51, name: 'Nina Sobczak', email: 'nina@fitmeet.dev', initials: 'NS' },
      { id: 52, name: 'Maya Kowalska', email: 'maya@fitmeet.dev', initials: 'MK' },
      { id: 53, name: 'Jan Tomasz', email: 'jan@fitmeet.dev', initials: 'JT' },
    ],
  },
];

const meetingsStorageKey = 'fit-the-meet-meetings-v1';
const groupsStorageKey = 'fit-the-meet-groups-v1';

function loadStoredMeetings() {
  try {
    const stored = JSON.parse(localStorage.getItem(meetingsStorageKey) || 'null');
    if (!Array.isArray(stored)) return initialMeetings;
    return stored.map((meeting: Meeting, index: number) => ({
      ...meeting,
      shareCode: meeting.shareCode || `${meeting.kind === 'poll' ? 'POLL' : 'EVENT'}-${meeting.id || index + 1}`,
    }));
  } catch {
    return initialMeetings;
  }
}

function loadStoredGroups() {
  try {
    const stored = JSON.parse(localStorage.getItem(groupsStorageKey) || 'null');
    return Array.isArray(stored) ? stored as WorkspaceGroup[] : initialGroups;
  } catch {
    return initialGroups;
  }
}

function Icon({ name, size = 20, stroke = 1.9 }: { name: IconName; size?: number; stroke?: number }) {
  const paths: Record<IconName, JSX.Element> = {
    calendar: <><rect x="3" y="5" width="18" height="16" rx="3"/><path d="M8 3v4M16 3v4M3 10h18"/></>,
    check: <><path d="m5 12 4 4L19 6"/><circle cx="12" cy="12" r="9"/></>,
    clock: <><circle cx="12" cy="12" r="9"/><path d="M12 7v5l3 2"/></>,
    filter: <><path d="M4 6h16M7 12h10M10 18h4"/></>,
    group: <><path d="M16 19v-1.5a4.5 4.5 0 0 0-4.5-4.5h-3A4.5 4.5 0 0 0 4 17.5V19"/><circle cx="10" cy="7" r="3"/><path d="M17 11a3 3 0 0 1 3 3v2M16 4.5a3 3 0 0 1 0 5"/></>,
    heart: <path d="M20.8 5.8a5.4 5.4 0 0 0-7.7 0L12 6.9l-1.1-1.1a5.4 5.4 0 0 0-7.7 7.7L12 22l8.8-8.5a5.4 5.4 0 0 0 0-7.7Z"/>,
    home: <><path d="m3 11 9-8 9 8"/><path d="M5 10v10h14V10M9 20v-6h6v6"/></>,
    lock: <><rect x="4" y="10" width="16" height="11" rx="3"/><path d="M8 10V7a4 4 0 0 1 8 0v3M12 14v3"/></>,
    location: <><path d="M20 10c0 5-8 11-8 11S4 15 4 10a8 8 0 1 1 16 0Z"/><circle cx="12" cy="10" r="2.5"/></>,
    mail: <><rect x="3" y="5" width="18" height="14" rx="3"/><path d="m4 7 8 6 8-6"/></>,
    plus: <path d="M12 5v14M5 12h14"/>,
    search: <><circle cx="11" cy="11" r="7"/><path d="m20 20-4-4"/></>,
    settings: <><path d="M4 7h10M18 7h2M4 17h2M10 17h10M14 5v4M8 15v4"/></>,
    sparkle: <><path d="m12 2 1.5 5.5L19 9l-5.5 1.5L12 16l-1.5-5.5L5 9l5.5-1.5L12 2Z"/><path d="m19 15 .7 2.3L22 18l-2.3.7L19 21l-.7-2.3L16 18l2.3-.7L19 15Z"/></>,
    user: <><circle cx="12" cy="8" r="4"/><path d="M4 21a8 8 0 0 1 16 0"/></>,
    video: <><rect x="3" y="6" width="13" height="12" rx="3"/><path d="m16 10 5-3v10l-5-3"/></>,
  };
  return (
    <svg width={size} height={size} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={stroke} strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
      {paths[name]}
    </svg>
  );
}

function AvailabilityGrid({
  dates,
  selected,
  onChange,
}: {
  dates: string[];
  selected: Set<string>;
  onChange: (updater: (current: Set<string>) => Set<string>) => void;
}) {
  const paintMode = useRef<boolean | null>(null);
  const lastPainted = useRef('');

  const paintSlot = (key: string) => {
    if (paintMode.current === null || lastPainted.current === key) return;
    const shouldSelect = paintMode.current;
    lastPainted.current = key;
    onChange((current) => {
      const next = new Set(current);
      if (shouldSelect) next.add(key);
      else next.delete(key);
      return next;
    });
  };

  const startPainting = (event: ReactPointerEvent<HTMLButtonElement>, key: string) => {
    if (event.button !== 0) return;
    event.preventDefault();
    paintMode.current = !selected.has(key);
    lastPainted.current = '';
    const grid = event.currentTarget.closest('[data-availability-grid]') as HTMLDivElement | null;
    grid?.setPointerCapture(event.pointerId);
    paintSlot(key);
  };

  const movePainting = (event: ReactPointerEvent<HTMLDivElement>) => {
    if (paintMode.current === null) return;
    const target = document.elementFromPoint(event.clientX, event.clientY) as HTMLElement | null;
    const cell = target?.closest<HTMLElement>('[data-slot]');
    if (cell?.dataset.slot) paintSlot(cell.dataset.slot);
  };

  const stopPainting = (event: ReactPointerEvent<HTMLDivElement>) => {
    if (event.currentTarget.hasPointerCapture(event.pointerId)) {
      event.currentTarget.releasePointerCapture(event.pointerId);
    }
    paintMode.current = null;
    lastPainted.current = '';
  };

  const toggleWithKeyboard = (event: React.KeyboardEvent<HTMLButtonElement>, key: string) => {
    if (event.key !== 'Enter' && event.key !== ' ') return;
    event.preventDefault();
    onChange((current) => {
      const next = new Set(current);
      if (next.has(key)) next.delete(key);
      else next.add(key);
      return next;
    });
  };

  return (
    <div className="availability-grid-scroll">
      <div
        className="availability-grid"
        data-availability-grid
        style={{ '--availability-days': dates.length } as React.CSSProperties}
        onPointerMove={movePainting}
        onPointerUp={stopPainting}
        onPointerCancel={stopPainting}
        role="grid"
        aria-label="Availability timetable"
      >
        <div className="availability-grid-corner" aria-hidden="true">Time</div>
        {dates.map((date) => {
          const [weekday, rest] = formatDay(date).split(', ');
          return (
            <div className="availability-day-heading" role="columnheader" key={date}>
              <span>{weekday}</span>
              <strong>{rest}</strong>
            </div>
          );
        })}
        {availabilityTimes.map((minutes) => (
          <div className={`availability-grid-row ${minutes % 60 === 0 ? 'is-hour' : ''}`} role="row" key={minutes}>
            <div className="availability-time-label" role="rowheader">
              {minutes % 60 === 0 ? minutesToTime(minutes) : ''}
            </div>
            {dates.map((date) => {
              const key = slotKey(date, minutes);
              const isSelected = selected.has(key);
              return (
                <button
                  className={`availability-cell ${isSelected ? 'is-selected' : ''}`}
                  type="button"
                  key={key}
                  data-slot={key}
                  aria-label={`${formatDay(date)}, ${minutesToTime(minutes)} to ${minutesToTime(minutes + slotMinutes)}`}
                  aria-pressed={isSelected}
                  onPointerDown={(event) => startPainting(event, key)}
                  onKeyDown={(event) => toggleWithKeyboard(event, key)}
                />
              );
            })}
          </div>
        ))}
      </div>
    </div>
  );
}

function AvailabilityResultsGrid({ meeting, dates }: { meeting: Meeting; dates: string[] }) {
  const counts = useMemo(() => {
    const result = new Map<string, number>();
    meeting.options.forEach((option) => {
      const block = availabilityFromOption(option);
      const start = Math.max(availabilityStartMinutes, timeToMinutes(block.start_time));
      const end = Math.min(availabilityEndMinutes, timeToMinutes(block.end_time));
      for (let minutes = start; minutes < end; minutes += slotMinutes) {
        result.set(slotKey(block.date, minutes), option.votes);
      }
    });
    return result;
  }, [meeting]);
  const maxVotes = Math.max(1, ...counts.values());
  const bestOption = [...meeting.options].sort((a, b) => b.votes - a.votes)[0];

  return (
    <Stack gap="md">
      {bestOption && (
        <Paper className="best-overlap" radius="lg">
          <Text size="xs" fw={900}>BEST OVERLAP</Text>
          <Text fw={850} mt={3}>{bestOption.label} · {bestOption.detail}</Text>
          <Text size="sm" c="dimmed" mt={2}>{bestOption.votes} of {meeting.participants} participants are available.</Text>
        </Paper>
      )}
      <div className="availability-grid-scroll">
        <div className="availability-grid results-grid" style={{ '--availability-days': dates.length } as React.CSSProperties} role="grid" aria-label="Availability results">
          <div className="availability-grid-corner" aria-hidden="true">Time</div>
          {dates.map((date) => {
            const [weekday, rest] = formatDay(date).split(', ');
            return <div className="availability-day-heading" role="columnheader" key={date}><span>{weekday}</span><strong>{rest}</strong></div>;
          })}
          {availabilityTimes.map((minutes) => (
            <div className={`availability-grid-row ${minutes % 60 === 0 ? 'is-hour' : ''}`} role="row" key={minutes}>
              <div className="availability-time-label" role="rowheader">{minutes % 60 === 0 ? minutesToTime(minutes) : ''}</div>
              {dates.map((date) => {
                const count = counts.get(slotKey(date, minutes)) || 0;
                const strength = count / maxVotes;
                return (
                  <div
                    className={`availability-result-cell ${count > 0 ? 'has-votes' : ''}`}
                    key={slotKey(date, minutes)}
                    role="gridcell"
                    aria-label={`${formatDay(date)}, ${minutesToTime(minutes)}: ${count} available`}
                    style={{ backgroundColor: count > 0 ? `rgba(111, 69, 237, ${0.12 + (strength * 0.78)})` : '#fff' }}
                  >
                    {count > 0 && <span>{count}</span>}
                  </div>
                );
              })}
            </div>
          ))}
        </div>
      </div>
      <Text size="xs" c="dimmed">Darker cells indicate more available participants.</Text>
    </Stack>
  );
}

function statusLabel(status: MeetingStatus) {
  return {
    'needs-response': 'Your response needed',
    collecting: 'Collecting availability',
    ready: 'Ready to schedule',
    scheduled: 'Scheduled',
  }[status];
}

type AuthMode = 'login' | 'register';
type Session = { kind: 'demo' } | { kind: 'api'; accessToken: string };

function errorMessage(error: unknown) {
  if (error instanceof ApiError) return error.message;
  return 'Something went wrong. Please try again.';
}

function AuthScreen({
  mode,
  onModeChange,
  onLogin,
  onRegister,
  onDemo,
}: {
  mode: AuthMode;
  onModeChange: (mode: AuthMode) => void;
  onLogin: (credentials: LoginCredentials, remember: boolean) => Promise<void>;
  onRegister: (details: RegistrationDetails) => Promise<void>;
  onDemo: () => void;
}) {
  const [error, setError] = useState('');
  const [notice, setNotice] = useState('');
  const [submitting, setSubmitting] = useState(false);
  const isLogin = mode === 'login';

  const submit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setError('');
    setNotice('');
    const form = new FormData(event.currentTarget);
    if (!isLogin && form.get('password') !== form.get('confirmPassword')) {
      setError('Passwords do not match.');
      return;
    }

    setSubmitting(true);
    try {
      const email = String(form.get('email'));
      const password = String(form.get('password'));
      if (isLogin) {
        await onLogin({ email, password }, form.get('remember') === 'on');
      } else {
        await onRegister({ username: String(form.get('username')), email, password });
        onModeChange('login');
        setNotice('Account created. Sign in with your new details.');
      }
    } catch (requestError) {
      setError(errorMessage(requestError));
    } finally {
      setSubmitting(false);
    }
  };

  const switchMode = (nextMode: AuthMode) => {
    setError('');
    setNotice('');
    onModeChange(nextMode);
  };

  return (
    <main className="auth-shell">
      <section className="auth-showcase">
        <Group gap={10} className="auth-brand">
          <div className="brand-mark"><Icon name="calendar" size={22} stroke={2.2} /></div>
          <div className="brand">fit<span>the</span>meet<i>.</i></div>
        </Group>

        <Box className="auth-message">
          <Text className="eyebrow">Meeting scheduler</Text>
          <Title order={1}>Find a time<br />for a meeting.</Title>
          <Text className="auth-lede">Create an event or poll and collect availability from participants.</Text>
        </Box>

        <Paper className="auth-preview-card" radius="lg">
          <Group justify="space-between" align="flex-start" wrap="nowrap">
            <Box>
              <Badge className="auth-preview-badge">Project group</Badge>
              <Title order={3} mt={12}>Project meeting</Title>
              <Text c="dimmed" size="sm" mt={4}>Select the times when you are available.</Text>
            </Box>
            <div className="response-stat"><strong>8/10</strong><span>responded</span></div>
          </Group>
          <Stack gap={9} mt={20}>
            <div className="auth-time-row is-best"><span><strong>Mon, 26 Oct</strong><small>10:00–10:50</small></span><Badge>8 free</Badge></div>
            <div className="auth-time-row"><span><strong>Tue, 27 Oct</strong><small>14:00–14:50</small></span><Badge color="gray">6 free</Badge></div>
          </Stack>
          <Group justify="space-between" mt={18}>
            <Avatar.Group spacing="sm">
              {['MK', 'AP', 'JT', 'LN'].map((person, index) => <Avatar key={person} color={['violet', 'pink', 'cyan', 'orange'][index]} size="sm">{person}</Avatar>)}
            </Avatar.Group>
            <Text size="sm" fw={800}>8 of 10 responses</Text>
          </Group>
        </Paper>

        <Group className="auth-trust" gap={10} wrap="nowrap">
          <ThemeIcon color="dark" radius="xl" size={34}><Icon name="check" size={17} /></ThemeIcon>
          <Text size="sm"><strong>Events, polls, and groups</strong><br />in one application.</Text>
        </Group>
      </section>

      <section className="auth-form-pane">
        <Group gap={9} className="auth-mobile-brand">
          <div className="brand-mark"><Icon name="calendar" size={20} stroke={2.2} /></div>
          <div className="brand">fit<span>the</span>meet<i>.</i></div>
        </Group>

        <Paper className="auth-card" radius="xl">
          <Text className="eyebrow">{isLogin ? 'Sign in' : 'Create account'}</Text>
          <Title order={2} mt={5}>{isLogin ? 'Open your events.' : 'Create an account.'}</Title>
          <Text c="dimmed" mt={9} size="sm">
            {isLogin ? 'Enter your email address and password.' : 'Enter your details to register.'}
          </Text>

          <form key={mode} onSubmit={submit} className="auth-form">
            <Stack gap="md">
              {!isLogin && (
                <TextInput
                  name="username"
                  label="Username"
                  placeholder="alexmorgan"
                  leftSection={<Icon name="user" size={18} />}
                  required
                  minLength={6}
                  autoComplete="username"
                />
              )}
              <TextInput
                name="email"
                type="email"
                label="Email address"
                placeholder="name@example.com"
                leftSection={<Icon name="mail" size={18} />}
                required
                autoComplete="email"
              />
              <PasswordInput
                name="password"
                label="Password"
                placeholder={isLogin ? 'Your password' : 'At least 8 characters'}
                leftSection={<Icon name="lock" size={18} />}
                required
                minLength={8}
                autoComplete={isLogin ? 'current-password' : 'new-password'}
              />
              {!isLogin && (
                <PasswordInput
                  name="confirmPassword"
                  label="Confirm password"
                  placeholder="Repeat your password"
                  leftSection={<Icon name="lock" size={18} />}
                  required
                  minLength={8}
                  autoComplete="new-password"
                />
              )}

              {isLogin ? (
                <Group justify="space-between" gap="sm">
                  <Checkbox name="remember" label="Keep me signed in" size="sm" />
                  <Anchor component="button" type="button" size="sm" fw={750} c="violet">Forgot password?</Anchor>
                </Group>
              ) : (
                <Checkbox
                  required
                  label={<Text size="sm">I agree to the <Anchor href="#" c="violet">Terms</Anchor> and <Anchor href="#" c="violet">Privacy Policy</Anchor>.</Text>}
                />
              )}

              {notice && <Text className="success-notice" role="status">{notice}</Text>}
              {error && <Text className="error-notice" role="alert">{error}</Text>}
              <Button type="submit" className="auth-submit" size="lg" loading={submitting}>
                {isLogin ? 'Sign in' : 'Create account'}
              </Button>
            </Stack>
          </form>

          <Divider label="or" labelPosition="center" my="xl" />
          <Button fullWidth variant="outline" color="dark" className="demo-button" onClick={onDemo}>
            Open demo
          </Button>

          <Text ta="center" size="sm" c="dimmed" mt="xl">
            {isLogin ? 'New to Fit the Meet?' : 'Already have an account?'}{' '}
            <Anchor component="button" type="button" fw={850} c="dark" onClick={() => switchMode(isLogin ? 'register' : 'login')}>
              {isLogin ? 'Create an account' : 'Sign in'}
            </Anchor>
          </Text>
        </Paper>
      </section>
    </main>
  );
}

function Dashboard({ session, onLogout }: { session: Session; onLogout: () => void }) {
  const [meetings, setMeetings] = useState<Meeting[]>(loadStoredMeetings);
  const [groups, setGroups] = useState<WorkspaceGroup[]>(loadStoredGroups);
  const [query, setQuery] = useState('');
  const [activeFilter, setActiveFilter] = useState('all');
  const [resourceFilter, setResourceFilter] = useState<'all' | 'event' | 'poll'>('all');
  const [accessFilter, setAccessFilter] = useState<'all' | 'open' | 'group'>('all');
  const [sort, setSort] = useState('urgent');
  const [selectedMeeting, setSelectedMeeting] = useState<Meeting | null>(null);
  const [detailView, setDetailView] = useState<'respond' | 'results'>('respond');
  const [managedEventId, setManagedEventId] = useState<number | string>('');
  const [availabilitySlots, setAvailabilitySlots] = useState<Set<string>>(() => new Set());
  const [selectedPollOptions, setSelectedPollOptions] = useState<Set<number>>(() => new Set());
  const [availabilityStartDate, setAvailabilityStartDate] = useState('2026-10-26');
  const [eventActive, setEventActive] = useState(true);
  const [operationPending, setOperationPending] = useState(false);
  const [notice, setNotice] = useState('');
  const [dashboardNotice, setDashboardNotice] = useState('');
  const [creationMode, setCreationMode] = useState<'standalone' | 'group'>('standalone');
  const [creationKind, setCreationKind] = useState<'event' | 'poll'>('event');
  const [creationPending, setCreationPending] = useState(false);
  const [creationError, setCreationError] = useState('');
  const [joinCodeError, setJoinCodeError] = useState('');
  const [meetingCodeCopied, setMeetingCodeCopied] = useState(false);
  const [meetingLinkCopied, setMeetingLinkCopied] = useState(false);
  const [groupAction, setGroupAction] = useState<'groups' | 'create' | 'join'>('groups');
  const [groupPending, setGroupPending] = useState(false);
  const [generatingGroupId, setGeneratingGroupId] = useState<number | null>(null);
  const [groupError, setGroupError] = useState('');
  const [activationCode, setActivationCode] = useState<{ code: string; groupId: number; groupName: string } | null>(null);
  const [activationCodeCopied, setActivationCodeCopied] = useState(false);
  const [managedGroupId, setManagedGroupId] = useState<number | null>(null);
  const [participantQuery, setParticipantQuery] = useState('');
  const [selectedParticipantIds, setSelectedParticipantIds] = useState<Set<number>>(() => new Set());
  const [participantPending, setParticipantPending] = useState(false);
  const [participantNotice, setParticipantNotice] = useState('');
  const [mobileFiltersOpened, mobileFilters] = useDisclosure(false);
  const [createOpened, createModal] = useDisclosure(false);
  const [joinOpened, joinModal] = useDisclosure(false);
  const [groupsOpened, groupsModal] = useDisclosure(false);
  const availabilityDates = useMemo(() => dateWindow(availabilityStartDate), [availabilityStartDate]);
  const availabilityBlocks = useMemo(() => blocksFromSlots(availabilitySlots), [availabilitySlots]);
  const managedGroup = groups.find((group) => group.id === managedGroupId) || null;
  const deepLinkHandled = useRef(false);
  const filteredParticipants = useMemo(() => {
    const normalizedQuery = participantQuery.trim().toLowerCase();
    if (!managedGroup || !normalizedQuery) return managedGroup?.participants || [];
    return managedGroup.participants.filter((participant) => (
      `${participant.name} ${participant.email}`.toLowerCase().includes(normalizedQuery)
    ));
  }, [managedGroup, participantQuery]);

  useEffect(() => {
    localStorage.setItem(meetingsStorageKey, JSON.stringify(meetings));
  }, [meetings]);

  useEffect(() => {
    localStorage.setItem(groupsStorageKey, JSON.stringify(groups));
  }, [groups]);

  const closeCreateModal = () => {
    setCreationError('');
    setCreationMode('standalone');
    setCreationKind('event');
    createModal.close();
  };

  const closeJoinModal = () => {
    setJoinCodeError('');
    joinModal.close();
  };

  const closeGroupsModal = () => {
    setGroupError('');
    setGroupAction('groups');
    setActivationCode(null);
    setActivationCodeCopied(false);
    setManagedGroupId(null);
    setParticipantQuery('');
    setSelectedParticipantIds(new Set());
    setParticipantNotice('');
    groupsModal.close();
  };

  const filteredMeetings = useMemo(() => {
    const normalizedQuery = query.trim().toLowerCase();
    const result = meetings.filter((meeting) => {
      const matchesQuery = !normalizedQuery || [meeting.title, meeting.description, meeting.group, meeting.organizer]
        .join(' ')
        .toLowerCase()
        .includes(normalizedQuery);
      const matchesResource = resourceFilter === 'all'
        || (resourceFilter === 'poll' && meeting.kind === 'poll')
        || (resourceFilter === 'event' && meeting.kind !== 'poll');
      const matchesAccess = accessFilter === 'all'
        || (accessFilter === 'group' && Boolean(meeting.groupId))
        || (accessFilter === 'open' && !meeting.groupId);
      const matchesFilter = activeFilter === 'all'
        || (activeFilter === 'needs-response' && meeting.status === 'needs-response')
        || (activeFilter === 'ready' && meeting.status === 'ready')
        || (activeFilter === 'favorites' && meeting.favorite)
        || activeFilter === 'this-week'
        || meeting.group === activeFilter;
      return matchesQuery && matchesResource && matchesAccess && matchesFilter;
    });
    return [...result].sort((a, b) => {
      if (sort === 'responses') return (b.responses / b.participants) - (a.responses / a.participants);
      if (sort === 'newest') return b.id - a.id;
      const priority: Record<MeetingStatus, number> = { 'needs-response': 0, collecting: 1, ready: 2, scheduled: 3 };
      return priority[a.status] - priority[b.status];
    });
  }, [accessFilter, activeFilter, meetings, query, resourceFilter, sort]);

  const openMeeting = (meeting: Meeting) => {
    const initialBlocks = meeting.options.map(availabilityFromOption);
    setSelectedMeeting(meeting);
    setManagedEventId(meeting.apiId === null ? '' : (meeting.apiId ?? meeting.id));
    setAvailabilitySlots(slotsFromBlocks(initialBlocks));
    setSelectedPollOptions(new Set());
    setAvailabilityStartDate(initialBlocks[0]?.date || '2026-10-26');
    setEventActive(meeting.isActive ?? true);
    setDetailView('respond');
    setNotice('');
    setMeetingCodeCopied(false);
    setMeetingLinkCopied(false);
    const url = new URL(window.location.href);
    url.searchParams.set('code', meeting.shareCode);
    window.history.replaceState({}, '', url);
  };

  const closeMeeting = () => {
    setSelectedMeeting(null);
    const url = new URL(window.location.href);
    url.searchParams.delete('code');
    window.history.replaceState({}, '', url);
  };

  useEffect(() => {
    if (deepLinkHandled.current) return;
    const code = new URLSearchParams(window.location.search).get('code')?.trim().toUpperCase();
    if (!code) {
      deepLinkHandled.current = true;
      return;
    }
    const meeting = meetings.find((item) => item.shareCode.toUpperCase() === code);
    deepLinkHandled.current = true;
    if (meeting) openMeeting(meeting);
    else {
      setJoinCodeError('No event or poll matches the code in this link.');
      joinModal.open();
    }
  }, [meetings]);

  const joinMeetingByCode = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    const code = String(new FormData(event.currentTarget).get('meetingCode') || '').trim().toUpperCase();
    const meeting = meetings.find((item) => item.shareCode.toUpperCase() === code);
    if (!meeting) {
      setJoinCodeError('No event or poll matches this code. Check the code and try again.');
      return;
    }
    setQuery('');
    setActiveFilter('all');
    closeJoinModal();
    openMeeting(meeting);
  };

  const copyMeetingCode = async () => {
    if (!selectedMeeting) return;
    try {
      await navigator.clipboard.writeText(selectedMeeting.shareCode);
      setMeetingCodeCopied(true);
    } catch {
      setNotice('Could not copy the code automatically. Select it and copy it manually.');
    }
  };

  const copyMeetingLink = async () => {
    if (!selectedMeeting) return;
    const url = new URL(window.location.href);
    url.searchParams.set('code', selectedMeeting.shareCode);
    try {
      await navigator.clipboard.writeText(url.toString());
      setMeetingLinkCopied(true);
    } catch {
      setNotice('Could not copy the link automatically. Copy it from the address bar.');
    }
  };

  const toggleFavorite = (id: number) => {
    setMeetings((current) => current.map((meeting) => meeting.id === id ? { ...meeting, favorite: !meeting.favorite } : meeting));
  };

  const parsedManagedEventId = typeof managedEventId === 'number' ? managedEventId : Number(managedEventId);
  const canParticipate = !selectedMeeting?.groupId || groups.some((group) => group.id === selectedMeeting.groupId);
  const selectedMeetingMaxVotes = selectedMeeting ? Math.max(1, ...selectedMeeting.options.map((option) => option.votes)) : 1;

  const saveAvailability = async () => {
    if (!selectedMeeting || !Number.isInteger(parsedManagedEventId) || parsedManagedEventId <= 0) {
      setNotice('Enter a valid event ID before saving availability.');
      return;
    }

    if (!eventActive) {
      setNotice('Activate the event before submitting availability.');
      return;
    }

    if (!canParticipate) {
      setNotice('Join this group before submitting availability.');
      return;
    }

    if (availabilityBlocks.length === 0) {
      setNotice('Add at least one availability block.');
      return;
    }

    setOperationPending(true);
    setNotice('');
    try {
      if (session.kind === 'api') {
        const payload = availabilityBlocks.map(({ date, start_time, end_time }) => ({ date, start_time, end_time }));
        await eventApi.saveAvailability(parsedManagedEventId, payload, session.accessToken);
      }
      setMeetings((current) => current.map((meeting) => meeting.id === selectedMeeting.id
        ? { ...meeting, status: 'collecting', responses: Math.min(meeting.responses + 1, meeting.participants) }
        : meeting));
      setSelectedMeeting((current) => current ? { ...current, status: 'collecting', responses: Math.min(current.responses + 1, current.participants) } : current);
      setNotice(session.kind === 'demo' ? 'Availability saved in demo mode.' : 'Availability saved to your API.');
    } catch (requestError) {
      setNotice(errorMessage(requestError));
    } finally {
      setOperationPending(false);
    }
  };

  const savePollAnswers = async () => {
    if (!selectedMeeting || !Number.isInteger(parsedManagedEventId) || parsedManagedEventId <= 0) {
      setNotice('Enter a valid poll ID before saving your choices.');
      return;
    }
    if (!eventActive) {
      setNotice('This poll is not active.');
      return;
    }
    if (!canParticipate) {
      setNotice('Join this group before responding to its poll.');
      return;
    }
    if (selectedPollOptions.size === 0) {
      setNotice('Choose at least one time option.');
      return;
    }

    setOperationPending(true);
    setNotice('');
    try {
      if (session.kind === 'api') {
        await pollApi.submitAnswers(parsedManagedEventId, [...selectedPollOptions], session.accessToken);
      }
      setMeetings((current) => current.map((meeting) => meeting.id === selectedMeeting.id
        ? {
          ...meeting,
          status: 'collecting',
          responses: Math.min(meeting.responses + 1, meeting.participants),
          options: meeting.options.map((option) => selectedPollOptions.has(option.id) ? { ...option, votes: option.votes + 1 } : option),
        }
        : meeting));
      setSelectedMeeting((current) => current ? {
        ...current,
        status: 'collecting',
        responses: Math.min(current.responses + 1, current.participants),
        options: current.options.map((option) => selectedPollOptions.has(option.id) ? { ...option, votes: option.votes + 1 } : option),
      } : current);
      setNotice(session.kind === 'demo' ? 'Poll choices saved in demo mode.' : 'Poll choices saved to your API.');
    } catch (requestError) {
      setNotice(errorMessage(requestError));
    } finally {
      setOperationPending(false);
    }
  };

  const toggleEventActive = async () => {
    if (!selectedMeeting || !Number.isInteger(parsedManagedEventId) || parsedManagedEventId <= 0) {
      setNotice('Enter a valid event ID before changing its status.');
      return;
    }

    if (eventActive && !window.confirm(`Deactivate this ${selectedMeeting.kind === 'poll' ? 'poll' : 'event'}? Participants will not be able to respond.`)) {
      return;
    }

    setOperationPending(true);
    setNotice('');
    try {
      if (session.kind === 'api') {
        if (selectedMeeting.kind === 'poll') await pollApi.toggleActive(parsedManagedEventId, session.accessToken);
        else await eventApi.toggleActive(parsedManagedEventId, session.accessToken);
      }
      const nextActive = !eventActive;
      setEventActive(nextActive);
      setMeetings((current) => current.map((meeting) => meeting.id === selectedMeeting.id
        ? { ...meeting, isActive: nextActive }
        : meeting));
      setNotice(`Event is now ${nextActive ? 'active' : 'inactive'}${session.kind === 'demo' ? ' in demo mode' : ''}.`);
    } catch (requestError) {
      setNotice(errorMessage(requestError));
    } finally {
      setOperationPending(false);
    }
  };

  const createMeeting = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    const form = new FormData(event.currentTarget);
    const title = String(form.get('title') || `Untitled ${creationKind}`);
    const durationMinutes = Number(form.get('duration') || 30);
    const groupId = creationMode === 'group' ? Number(form.get('groupId')) : undefined;
    if (creationMode === 'group' && (!Number.isInteger(groupId) || Number(groupId) <= 0)) {
      setCreationError(`Choose a group you own before creating this ${creationKind}.`);
      return;
    }

    const pollOptions = creationKind === 'poll'
      ? [1, 2].map((index) => ({
        start_time: String(form.get(`option${index}Start`) || ''),
        end_time: String(form.get(`option${index}End`) || ''),
      }))
      : [];
    if (creationKind === 'poll' && pollOptions.some((option) => !option.start_time || !option.end_time || option.start_time >= option.end_time)) {
      setCreationError('Each poll option needs an end time after its start time.');
      return;
    }

    setCreationPending(true);
    setCreationError('');
    try {
      if (session.kind === 'api') {
        const input = {
          name: title,
          description: String(form.get('description') || ''),
          duration_minutes: durationMinutes,
        };
        if (creationKind === 'poll') {
          await pollApi.create({ ...input, options: pollOptions }, session.accessToken, groupId);
        } else {
          await eventApi.create(input, session.accessToken, groupId);
        }
      }

      const nextId = Math.max(...meetings.map((meeting) => meeting.id)) + 1;
      const targetGroup = groups.find((group) => group.id === groupId);
      const group = targetGroup?.name || 'Open event';
      const nextOptions = creationKind === 'poll'
        ? pollOptions.map((option, index) => {
          const start = new Date(option.start_time);
          return {
            id: Date.now() + index,
            label: new Intl.DateTimeFormat('en-GB', { weekday: 'long', day: 'numeric', month: 'short' }).format(start),
            detail: `${option.start_time.slice(11, 16)}–${option.end_time.slice(11, 16)}`,
            votes: 0,
          };
        })
        : [
          { id: Date.now(), label: 'Monday, 2 Nov', detail: '10:00–10:30', votes: 0 },
          { id: Date.now() + 1, label: 'Tuesday, 3 Nov', detail: '14:00–14:30', votes: 0 },
        ];
      const next: Meeting = {
        id: nextId,
        shareCode: `${creationKind === 'poll' ? 'POLL' : 'EVENT'}-${nextId}-${Math.random().toString(36).slice(2, 6).toUpperCase()}`,
        title,
        description: String(form.get('description') || `New ${creationKind}.`),
        group,
        location: 'Remote',
        organizer: 'You',
        dateLabel: 'Just created',
        duration: `${durationMinutes} min`,
        responses: 0,
        participants: creationMode === 'group' ? 5 : 1,
        status: 'collecting',
        accent: '#6f45ed',
        softAccent: '#f3efff',
        icon: creationKind === 'poll' ? 'clock' : 'calendar',
        attendees: ['YOU'],
        isActive: true,
        apiId: session.kind === 'demo' ? nextId : null,
        kind: creationKind,
        groupId,
        options: nextOptions,
      };
      setMeetings((current) => [next, ...current]);
      setActiveFilter('all');
      setDashboardNotice(session.kind === 'demo'
        ? `${creationKind === 'poll' ? 'Poll' : 'Event'} created in demo mode.`
        : `${creationKind === 'poll' ? 'Poll' : 'Event'} created through your API. The create response does not include its new ID.`);
      closeCreateModal();
    } catch (requestError) {
      setCreationError(errorMessage(requestError));
    } finally {
      setCreationPending(false);
    }
  };

  const createGroup = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    const formElement = event.currentTarget;
    const form = new FormData(formElement);
    const name = String(form.get('groupName') || '').trim();
    const providedID = Number(form.get('groupId'));
    const groupID = Number.isInteger(providedID) && providedID > 0
      ? providedID
      : Math.max(100, ...groups.map((group) => group.id)) + 1;
    if (!name || (session.kind === 'api' && (!Number.isInteger(providedID) || providedID <= 0))) {
      setGroupError(session.kind === 'api' ? 'Enter the group name and its numeric API ID.' : 'Enter a group name.');
      return;
    }

    setGroupPending(true);
    setGroupError('');
    try {
      if (session.kind === 'api') await groupApi.create(name, session.accessToken);
      setGroups((current) => current.some((group) => group.id === groupID)
        ? current
        : [...current, { id: groupID, name, color: '#6f45ed', role: 'owner', participants: [] }]);
      setDashboardNotice(`Group “${name}” created. You can now add events and polls to it.`);
      formElement.reset();
      setGroupAction('groups');
    } catch (requestError) {
      setGroupError(errorMessage(requestError));
    } finally {
      setGroupPending(false);
    }
  };

  const joinGroup = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    const formElement = event.currentTarget;
    const form = new FormData(formElement);
    const code = String(form.get('activationCode') || '').trim();
    const name = String(form.get('joinedGroupName') || '').trim();
    const groupID = Number(form.get('joinedGroupId'));
    if (!code || !name || !Number.isInteger(groupID) || groupID <= 0) {
      setGroupError('Enter the activation code, group name, and numeric group ID.');
      return;
    }

    setGroupPending(true);
    setGroupError('');
    try {
      if (session.kind === 'api') await groupApi.join(code, session.accessToken);
      setGroups((current) => current.some((group) => group.id === groupID)
        ? current
        : [...current, { id: groupID, name, color: '#168a56', role: 'member', participants: [] }]);
      setDashboardNotice(`You joined “${name}”. Group events and polls are now available.`);
      formElement.reset();
      setGroupAction('groups');
    } catch (requestError) {
      setGroupError(errorMessage(requestError));
    } finally {
      setGroupPending(false);
    }
  };

  const generateActivationCode = async (group: WorkspaceGroup) => {
    setGeneratingGroupId(group.id);
    setGroupError('');
    setActivationCode(null);
    setActivationCodeCopied(false);
    try {
      const code = session.kind === 'api'
        ? (await groupApi.generateActivationCode(group.id, session.accessToken)).code
        : `FTM-${group.id}-${Math.random().toString(36).slice(2, 8).toUpperCase()}`;
      setActivationCode({ code, groupId: group.id, groupName: group.name });
    } catch (requestError) {
      setGroupError(errorMessage(requestError));
    } finally {
      setGeneratingGroupId(null);
    }
  };

  const copyActivationCode = async () => {
    if (!activationCode) return;
    try {
      await navigator.clipboard.writeText(activationCode.code);
      setActivationCodeCopied(true);
    } catch {
      setGroupError('Could not copy automatically. Select the code and copy it manually.');
    }
  };

  const openParticipantManager = (group: WorkspaceGroup) => {
    setManagedGroupId(group.id);
    setParticipantQuery('');
    setSelectedParticipantIds(new Set());
    setParticipantNotice('');
    setGroupError('');
  };

  const closeParticipantManager = () => {
    setManagedGroupId(null);
    setParticipantQuery('');
    setSelectedParticipantIds(new Set());
    setParticipantNotice('');
  };

  const toggleParticipant = (participantId: number) => {
    setSelectedParticipantIds((current) => {
      const next = new Set(current);
      if (next.has(participantId)) next.delete(participantId);
      else next.add(participantId);
      return next;
    });
  };

  const toggleAllVisibleParticipants = () => {
    const visibleIDs = filteredParticipants.map((participant) => participant.id);
    const allVisibleSelected = visibleIDs.length > 0 && visibleIDs.every((id) => selectedParticipantIds.has(id));
    setSelectedParticipantIds((current) => {
      const next = new Set(current);
      visibleIDs.forEach((id) => {
        if (allVisibleSelected) next.delete(id);
        else next.add(id);
      });
      return next;
    });
  };

  const removeSelectedParticipants = async () => {
    if (!managedGroup || managedGroup.role !== 'owner' || selectedParticipantIds.size === 0) return;
    const idsToRemove = [...selectedParticipantIds];
    if (!window.confirm(`Remove ${idsToRemove.length} ${idsToRemove.length === 1 ? 'member' : 'members'} from ${managedGroup.name}?`)) return;
    setParticipantPending(true);
    setParticipantNotice('');
    try {
      if (session.kind === 'api') {
        await groupApi.removeParticipants(managedGroup.id, idsToRemove, session.accessToken);
      }
      setGroups((current) => current.map((group) => group.id === managedGroup.id
        ? { ...group, participants: group.participants.filter((participant) => !selectedParticipantIds.has(participant.id)) }
        : group));
      setSelectedParticipantIds(new Set());
      setParticipantNotice(`${idsToRemove.length} ${idsToRemove.length === 1 ? 'participant' : 'participants'} removed${session.kind === 'demo' ? ' in demo mode' : ''}.`);
    } catch (requestError) {
      setParticipantNotice(errorMessage(requestError));
    } finally {
      setParticipantPending(false);
    }
  };

  const countFor = (id: string) => {
    if (id === 'all' || id === 'this-week') return meetings.length;
    if (id === 'needs-response') return meetings.filter((meeting) => meeting.status === 'needs-response').length;
    if (id === 'ready') return meetings.filter((meeting) => meeting.status === 'ready').length;
    if (id === 'favorites') return meetings.filter((meeting) => meeting.favorite).length;
    return meetings.filter((meeting) => meeting.group === id).length;
  };

  const SidebarContent = () => (
    <Stack gap={30}>
      <Box>
        <Text className="sidebar-kicker">My schedule</Text>
        <Stack gap={5} mt={12}>
          {filterItems.map((item) => (
            <UnstyledButton
              key={item.id}
              className={`sidebar-item ${activeFilter === item.id ? 'is-active' : ''}`}
              onClick={() => { setActiveFilter(item.id); mobileFilters.close(); }}
            >
              <span className="sidebar-icon" style={{ color: item.color }}><Icon name={item.icon} size={21} /></span>
              <span>{item.label}</span>
              <span className="sidebar-count">{countFor(item.id)}</span>
            </UnstyledButton>
          ))}
        </Stack>
      </Box>
      <Box>
        <Group justify="space-between" gap="xs">
          <Text className="sidebar-kicker">Groups</Text>
          <ActionIcon variant="subtle" color="dark" size="sm" aria-label="Manage groups" onClick={groupsModal.open}>
            <Icon name="plus" size={16} />
          </ActionIcon>
        </Group>
        <Stack gap={5} mt={12}>
          {groups.map((item) => (
            <UnstyledButton
              key={item.id}
              className={`sidebar-item ${activeFilter === item.name ? 'is-active' : ''}`}
              onClick={() => { setActiveFilter(item.name); mobileFilters.close(); }}
            >
              <span className="group-dot" style={{ background: item.color }} />
              <span>{item.name}</span>
              <span className="sidebar-count">{countFor(item.name)}</span>
            </UnstyledButton>
          ))}
          <Button variant="subtle" color="dark" size="compact-sm" leftSection={<Icon name="group" size={15} />} onClick={groupsModal.open}>
            Create or join group
          </Button>
        </Stack>
      </Box>
      <Paper className="sidebar-tip" radius="xl">
        <ThemeIcon color="dark" variant="filled" radius="xl" size={34}><Icon name="sparkle" size={17} /></ThemeIcon>
        <Text fw={800} mt={12}>How it works</Text>
        <Text size="sm" c="dimmed" mt={5}>Create an event, share it, and collect responses.</Text>
      </Paper>
    </Stack>
  );

  return (
    <Box className="app-shell">
      <header className="topbar">
        <Group className="brand-wrap" gap={10} wrap="nowrap">
          <Burger opened={mobileFiltersOpened} onClick={mobileFilters.toggle} hiddenFrom="md" size="sm" aria-label="Open event filters" />
          <div className="brand-mark"><Icon name="calendar" size={22} stroke={2.2} /></div>
          <div className="brand">fit<span>the</span>meet<i>.</i></div>
        </Group>

        <Paper className="search-shell" radius="xl">
          <TextInput
            value={query}
            onChange={(event) => setQuery(event.currentTarget.value)}
            placeholder="Search events, polls, groups, or people…"
            variant="unstyled"
            aria-label="Search"
            leftSection={<Icon name="search" size={20} />}
          />
          <Divider orientation="vertical" />
          <Button className="workspace-button" variant="subtle" color="dark" leftSection={<Icon name="settings" size={18} />}>Settings</Button>
        </Paper>

        <Group gap={10} wrap="nowrap">
          {session.kind === 'demo' && <Badge className="mode-badge">Demo</Badge>}
          <Button className="join-button" variant="default" leftSection={<Icon name="lock" size={17} />} onClick={joinModal.open}>Join by code</Button>
          <Button className="create-button" leftSection={<Icon name="plus" size={19} />} onClick={createModal.open}>New</Button>
          <Menu position="bottom-end" shadow="md" width={210}>
            <Menu.Target>
              <UnstyledButton className="avatar-button" aria-label="Open account menu">
                <Avatar color="dark" radius="xl">AK</Avatar>
              </UnstyledButton>
            </Menu.Target>
            <Menu.Dropdown>
              <Menu.Label>Signed in as Alex</Menu.Label>
              <Menu.Item leftSection={<Icon name="user" size={17} />} onClick={onLogout}>Sign out</Menu.Item>
            </Menu.Dropdown>
          </Menu>
        </Group>
      </header>

      <main className="workspace-layout">
        <aside className="desktop-sidebar"><SidebarContent /></aside>

        <section className="content-column">
          <Group justify="space-between" align="flex-end" className="content-heading">
            <Box>
              <Text className="eyebrow">Sunday, 25 October</Text>
              <Title order={1}>Events and polls</Title>
              <Text c="dimmed" mt={6}>
                <strong>{countFor('needs-response')} {countFor('needs-response') === 1 ? 'event needs' : 'events need'}</strong> your availability.
              </Text>
            </Box>
            <Group className="dashboard-controls" gap="sm" wrap="wrap" justify="flex-end">
              <Select
                value={resourceFilter}
                onChange={(value) => setResourceFilter((value || 'all') as 'all' | 'event' | 'poll')}
                data={[
                  { value: 'all', label: 'Events and polls' },
                  { value: 'event', label: 'Events only' },
                  { value: 'poll', label: 'Polls only' },
                ]}
                allowDeselect={false}
                aria-label="Filter by type"
              />
              <Select
                value={accessFilter}
                onChange={(value) => setAccessFilter((value || 'all') as 'all' | 'open' | 'group')}
                data={[
                  { value: 'all', label: 'All access' },
                  { value: 'open', label: 'Open access' },
                  { value: 'group', label: 'Group only' },
                ]}
                allowDeselect={false}
                aria-label="Filter by access"
              />
              <Select
                className="sort-select"
                value={sort}
                onChange={(value) => setSort(value || 'urgent')}
                data={[
                  { value: 'urgent', label: 'Needs response first' },
                  { value: 'newest', label: 'Newest first' },
                  { value: 'responses', label: 'Most responses' },
                ]}
                allowDeselect={false}
                leftSection={<Icon name="filter" size={17} />}
                aria-label="Sort events"
              />
            </Group>
          </Group>

          {dashboardNotice && (
            <Text className="success-notice dashboard-notice" role="status" mt={18}>
              {dashboardNotice}
              <button type="button" onClick={() => setDashboardNotice('')} aria-label="Dismiss notification">×</button>
            </Text>
          )}

          {session.kind === 'api' && (
            <Paper className="api-scope-note" radius="lg" mt={18}>
              <ThemeIcon color="violet" variant="light" radius="xl"><Icon name="sparkle" size={16} /></ThemeIcon>
              <Text size="sm"><strong>API mode.</strong> The API does not provide an event list, so the example events remain visible. Enter an event ID before updating it.</Text>
            </Paper>
          )}

          <Stack gap={21} mt={28}>
            {filteredMeetings.map((meeting) => (
              <article
                key={meeting.id}
                className="meeting-card"
                style={{ '--accent': meeting.accent, '--soft-accent': meeting.softAccent } as React.CSSProperties}
                onClick={() => openMeeting(meeting)}
              >
                <Group justify="space-between" align="flex-start" wrap="nowrap" gap="xl">
                  <Box className="meeting-copy">
                    <Group gap={10} wrap="nowrap">
                      <Title order={2}>{meeting.title}</Title>
                      {meeting.kind === 'poll' && <Badge className="resource-kind-badge" variant="light">Poll</Badge>}
                      {meeting.groupId && <Badge className="scope-badge" variant="outline" leftSection={<Icon name="lock" size={12} />}>Group</Badge>}
                      {meeting.status === 'needs-response' && <span className="attention-dot" title="Needs your response" />}
                      {meeting.isActive === false && <Badge color="gray" variant="light" size="sm">Inactive</Badge>}
                    </Group>
                    <Text className="meeting-description">{meeting.description}</Text>
                  </Box>
                  <Box className="response-stat">
                    <strong>{meeting.responses}/{meeting.participants}</strong>
                    <span>responded</span>
                  </Box>
                </Group>

                <Group justify="space-between" align="center" mt={18} wrap="nowrap" className="meeting-footer">
                  <Group gap={13} className="meeting-meta">
                    <Badge className="status-badge" leftSection={<Icon name={meeting.icon} size={15} />}>{statusLabel(meeting.status)}</Badge>
                    <span className="meta-item"><Icon name={meeting.location === 'Google Meet' || meeting.location === 'Remote' ? 'video' : 'location'} size={18} />{meeting.location}</span>
                    <span className="meta-separator">•</span>
                    <span className="meta-item"><Icon name="user" size={18} />{meeting.organizer}</span>
                  </Group>
                  <Group gap={14} wrap="nowrap" className="meeting-actions">
                    <span className="deadline"><Icon name="clock" size={18} />{meeting.dateLabel}</span>
                    <ActionIcon
                      className={`favorite-button ${meeting.favorite ? 'is-favorite' : ''}`}
                      variant="transparent"
                      aria-label={meeting.favorite ? 'Remove from favorites' : 'Add to favorites'}
                      onClick={(event) => { event.stopPropagation(); toggleFavorite(meeting.id); }}
                    >
                      <Icon name="heart" size={22} />
                    </ActionIcon>
                  </Group>
                </Group>
              </article>
            ))}
          </Stack>

          {filteredMeetings.length === 0 && (
            <Paper className="empty-state" radius="xl">
              <ThemeIcon color="violet" variant="light" radius="xl" size={52}><Icon name="search" /></ThemeIcon>
              <Title order={3} mt={16}>No events found</Title>
              <Text c="dimmed" mt={5}>Try a different search or filter.</Text>
              <Button variant="light" color="dark" mt={18} onClick={() => { setQuery(''); setActiveFilter('all'); setResourceFilter('all'); setAccessFilter('all'); }}>Clear filters</Button>
            </Paper>
          )}
        </section>
      </main>

      <Drawer opened={mobileFiltersOpened} onClose={mobileFilters.close} title={<span className="drawer-title">Browse events</span>} size="xs" padding="lg">
        <SidebarContent />
      </Drawer>

      <Drawer
        opened={Boolean(selectedMeeting)}
        onClose={closeMeeting}
        position="right"
        size={820}
        padding="xl"
        title={<span className="drawer-title">{selectedMeeting?.kind === 'poll' ? 'Poll details' : 'Event details'}</span>}
        overlayProps={{ backgroundOpacity: 0.18, blur: 2 }}
      >
        {selectedMeeting && (
          <Stack gap={24}>
            <Box>
              <Badge className="drawer-badge" style={{ background: selectedMeeting.softAccent, color: selectedMeeting.accent }}>
                {selectedMeeting.group}
              </Badge>
              <Title order={2} mt={12}>{selectedMeeting.title}</Title>
              <Text c="dimmed" mt={8}>{selectedMeeting.description}</Text>
            </Box>

            <Group gap={18}>
              <span className="meta-item dark"><Icon name="clock" size={18} />{selectedMeeting.duration}</span>
              <span className="meta-item dark"><Icon name="group" size={18} />{selectedMeeting.participants} invited</span>
              <Badge variant="light" color={selectedMeeting.groupId ? 'violet' : 'green'}>
                {selectedMeeting.groupId ? 'Group only' : 'Open access'}
              </Badge>
            </Group>

            <Paper className="meeting-code-panel" radius="lg">
              <Group justify="space-between" align="center" wrap="wrap">
                <Box>
                  <Text size="xs" fw={900} c="dimmed">JOIN CODE</Text>
                  <Text component="code">{selectedMeeting.shareCode}</Text>
                  <Text size="xs" c="dimmed" mt={3}>
                    {selectedMeeting.groupId
                      ? 'This code opens the item. Group membership is still required to respond.'
                      : 'Enter this code from the main screen to open this item.'}
                  </Text>
                </Box>
                <Group gap={6} wrap="nowrap">
                  <Button size="xs" variant="default" onClick={copyMeetingCode}>{meetingCodeCopied ? 'Code copied' : 'Copy code'}</Button>
                  <Button size="xs" variant="light" color={meetingLinkCopied ? 'green' : 'violet'} onClick={copyMeetingLink}>
                    {meetingLinkCopied ? 'Link copied' : 'Copy link'}
                  </Button>
                </Group>
              </Group>
            </Paper>

            {selectedMeeting.groupId && !canParticipate && (
              <Paper className="membership-warning" radius="lg">
                <ThemeIcon color="orange" variant="light" radius="xl"><Icon name="lock" size={16} /></ThemeIcon>
                <Box>
                  <Text fw={850} size="sm">Members only</Text>
                  <Text c="dimmed" size="xs">Join {selectedMeeting.group} with an activation code before responding.</Text>
                </Box>
                <Button size="xs" variant="light" color="orange" onClick={groupsModal.open}>Join group</Button>
              </Paper>
            )}

            <Paper className="event-control-panel" radius="xl">
              <Group justify="space-between" align="flex-start" wrap="nowrap">
                <Box>
                  <Text fw={850}>API {selectedMeeting.kind === 'poll' ? 'poll' : 'event'}</Text>
                  <Text size="xs" c="dimmed" mt={2}>Used by status and availability actions.</Text>
                </Box>
                <Badge color={eventActive ? 'green' : 'gray'} variant="light">{eventActive ? 'Active' : 'Inactive'}</Badge>
              </Group>
              <NumberInput
                label={selectedMeeting.kind === 'poll' ? 'Poll ID' : 'Event ID'}
                value={managedEventId}
                onChange={setManagedEventId}
                min={1}
                allowDecimal={false}
                placeholder="e.g. 42"
                mt="md"
              />
              <Button
                fullWidth
                mt="md"
                variant="outline"
                color={eventActive ? 'red' : 'green'}
                loading={operationPending}
                onClick={toggleEventActive}
              >
                {session.kind === 'api' ? 'Toggle active status' : (eventActive ? 'Deactivate event' : 'Activate event')}
              </Button>
            </Paper>

            {canParticipate && (
              <SegmentedControl
                className="detail-view-switch"
                fullWidth
                value={detailView}
                onChange={(value) => setDetailView(value as 'respond' | 'results')}
                data={[
                  { value: 'respond', label: selectedMeeting.kind === 'poll' ? 'Your response' : 'Your availability' },
                  { value: 'results', label: 'Results' },
                ]}
              />
            )}

            {canParticipate && selectedMeeting.kind === 'poll' && detailView === 'respond' && (
              <Box>
                <Title order={3}>Choose the times that work</Title>
                <Text size="sm" c="dimmed" mt={4}>Select one or more options, then save your response.</Text>
                <Stack gap={10} mt={15}>
                  {selectedMeeting.options.map((option) => {
                    const isSelected = selectedPollOptions.has(option.id);
                    return (
                      <UnstyledButton
                        key={option.id}
                        className={`time-option ${isSelected ? 'is-selected' : ''}`}
                        aria-pressed={isSelected}
                        onClick={() => setSelectedPollOptions((current) => {
                          const next = new Set(current);
                          if (next.has(option.id)) next.delete(option.id);
                          else next.add(option.id);
                          return next;
                        })}
                      >
                        <span className="option-check">{isSelected && <Icon name="check" size={16} />}</span>
                        <span className="option-copy"><strong>{option.label}</strong><small>{option.detail}</small></span>
                        <span className="option-votes">{option.votes} free</span>
                      </UnstyledButton>
                    );
                  })}
                </Stack>
              </Box>
            )}

            {canParticipate && selectedMeeting.kind === 'poll' && detailView === 'results' && (
              <Box>
                <Title order={3}>Poll results</Title>
                <Text size="sm" c="dimmed" mt={4}>{selectedMeeting.responses} of {selectedMeeting.participants} participants responded.</Text>
                <Stack gap="sm" mt="md">
                  {[...selectedMeeting.options].sort((a, b) => b.votes - a.votes).map((option, index) => (
                    <Paper className="poll-result-row" radius="lg" key={option.id}>
                      <Group justify="space-between" gap="md" wrap="nowrap">
                        <Box className="option-copy">
                          <strong>{option.label}</strong>
                          <small>{option.detail}</small>
                        </Box>
                        <Group gap={7} wrap="nowrap">
                          {index === 0 && <Badge size="xs" color="violet" variant="light">Best</Badge>}
                          <Text size="sm" fw={850}>{option.votes}</Text>
                        </Group>
                      </Group>
                      <div className="poll-result-track" aria-label={`${option.votes} votes`}>
                        <span style={{ width: `${(option.votes / selectedMeetingMaxVotes) * 100}%` }} />
                      </div>
                    </Paper>
                  ))}
                </Stack>
              </Box>
            )}

            {canParticipate && selectedMeeting.kind !== 'poll' && detailView === 'respond' && (
              <Box>
                <Group justify="space-between" align="flex-start">
                  <Box>
                    <Title order={3}>Your availability</Title>
                    <Text size="sm" c="dimmed" mt={4}>Drag across the grid to select the times when you are available.</Text>
                  </Box>
                  <Button size="xs" variant="subtle" color="gray" disabled={availabilitySlots.size === 0} onClick={() => setAvailabilitySlots(new Set())}>
                    Clear all
                  </Button>
                </Group>
                <Paper className="availability-picker" radius="xl" mt={15}>
                  <Group justify="space-between" align="center" gap="sm" mb="md">
                    <Group gap={7} wrap="nowrap">
                      <ActionIcon variant="default" radius="md" aria-label="Show previous five days" onClick={() => setAvailabilityStartDate((current) => addDays(current, -5))}>
                        <span aria-hidden="true">‹</span>
                      </ActionIcon>
                      <Text className="availability-range-label">
                        {formatDay(availabilityDates[0])} – {formatDay(availabilityDates[availabilityDates.length - 1])}
                      </Text>
                      <ActionIcon variant="default" radius="md" aria-label="Show next five days" onClick={() => setAvailabilityStartDate((current) => addDays(current, 5))}>
                        <span aria-hidden="true">›</span>
                      </ActionIcon>
                    </Group>
                    <Badge className="availability-total" variant="light">
                      {availabilitySlots.size === 0 ? 'Nothing selected' : `${formatSelectedDuration(availabilitySlots.size)} selected`}
                    </Badge>
                  </Group>

                  <AvailabilityGrid dates={availabilityDates} selected={availabilitySlots} onChange={setAvailabilitySlots} />

                  <Group className="availability-legend" gap="lg" mt="md">
                    <span><i className="legend-swatch is-free" />Available</span>
                    <span><i className="legend-swatch" />Not available</span>
                    <Text size="xs" c="dimmed">Start dragging on a purple cell to erase.</Text>
                  </Group>
                </Paper>

                {availabilityBlocks.length > 0 && (
                  <div className="availability-summary" aria-live="polite">
                    <Text size="xs" fw={900} c="dimmed">SELECTED RANGES</Text>
                    <Group gap={7} mt={8}>
                      {availabilityBlocks.slice(0, 5).map((block) => (
                        <Badge variant="outline" color="dark" key={`${block.date}-${block.start_time}`}>
                          {formatDay(block.date)} · {block.start_time}–{block.end_time}
                        </Badge>
                      ))}
                      {availabilityBlocks.length > 5 && <Badge variant="light" color="gray">+{availabilityBlocks.length - 5} more</Badge>}
                    </Group>
                  </div>
                )}
              </Box>
            )}

            {canParticipate && selectedMeeting.kind !== 'poll' && detailView === 'results' && (
              <Box>
                <Title order={3}>Availability results</Title>
                <Text size="sm" c="dimmed" mt={4} mb="md">Compare participant availability across the selected dates.</Text>
                <AvailabilityResultsGrid meeting={selectedMeeting} dates={availabilityDates} />
              </Box>
            )}

            {notice && <Text role="status" className={/(saved|now active|now inactive)/i.test(notice) ? 'success-notice' : 'error-notice'}>{notice}</Text>}
            {detailView === 'respond' && (
              <Button
                className="save-vote-button"
                size="lg"
                loading={operationPending}
                disabled={!eventActive || !canParticipate}
                onClick={selectedMeeting.kind === 'poll' ? savePollAnswers : saveAvailability}
              >
                {selectedMeeting.kind === 'poll' ? 'Save poll choices' : 'Save availability'}
              </Button>
            )}
          </Stack>
        )}
      </Drawer>

      <Modal opened={joinOpened} onClose={closeJoinModal} title={<span className="modal-title">Join by code</span>} centered size="md" padding="xl" radius="xl">
        <form onSubmit={joinMeetingByCode}>
          <Stack gap="md">
            <Text c="dimmed" size="sm">Enter an event or poll code.</Text>
            <TextInput
              name="meetingCode"
              label="Code"
              placeholder="EVENT-DEMO"
              leftSection={<Icon name="lock" size={17} />}
              onChange={() => setJoinCodeError('')}
              autoFocus
              required
            />
            <Paper className="demo-code-note" radius="lg">
              <Text size="xs" fw={900}>DEMO CODES</Text>
              <Text size="sm" mt={4}><code>EVENT-DEMO</code> · availability event</Text>
              <Text size="sm" mt={2}><code>POLL-DEMO</code> · time poll</Text>
            </Paper>
            {joinCodeError && <Text className="error-notice" role="alert">{joinCodeError}</Text>}
            <Group justify="flex-end" mt="xs">
              <Button variant="subtle" color="dark" onClick={closeJoinModal}>Cancel</Button>
              <Button type="submit" className="create-button">Open</Button>
            </Group>
          </Stack>
        </form>
      </Modal>

      <Modal opened={createOpened} onClose={closeCreateModal} title={<span className="modal-title">Create event or poll</span>} centered size="lg" padding="xl" radius="xl">
        <form onSubmit={createMeeting}>
          <Stack gap="md">
            <Text c="dimmed" size="sm">Choose a type and decide whether participation requires group membership.</Text>
            <SegmentedControl
              fullWidth
              value={creationKind}
              onChange={(value) => { setCreationKind(value as 'event' | 'poll'); setCreationError(''); }}
              data={[
                { value: 'event', label: 'Availability grid' },
                { value: 'poll', label: 'Time poll' },
              ]}
            />
            <SegmentedControl
              fullWidth
              value={creationMode}
              onChange={(value) => { setCreationMode(value as 'standalone' | 'group'); setCreationError(''); }}
              data={[
                { value: 'standalone', label: 'Outside a group' },
                { value: 'group', label: 'Inside a group' },
              ]}
            />
            <TextInput name="title" label={creationKind === 'poll' ? 'Poll name' : 'Event name'} placeholder="e.g. Weekly meeting" required minLength={6} />
            <Textarea name="description" label="Description" placeholder="Add a short description" minRows={3} />
            {creationMode === 'group' && (
              <Select
                name="groupId"
                label="Group"
                placeholder="Choose a group you own"
                data={groups.filter((group) => group.role === 'owner').map((group) => ({ value: String(group.id), label: `${group.name} · #${group.id}` }))}
                required
                searchable
                nothingFoundMessage="Create a group first"
              />
            )}
            <NumberInput name="duration" label="Duration (minutes)" defaultValue={30} min={1} max={1440} step={15} required />
            {creationKind === 'poll' && (
              <Stack gap="sm">
                <Text size="sm" fw={850}>Time options</Text>
                {[1, 2].map((index) => (
                  <Paper className="poll-option-editor" radius="lg" key={index}>
                    <Text size="xs" fw={900} c="dimmed">OPTION {index}</Text>
                    <Group grow mt="xs">
                      <TextInput name={`option${index}Start`} type="datetime-local" label="Starts" required />
                      <TextInput name={`option${index}End`} type="datetime-local" label="Ends" required />
                    </Group>
                  </Paper>
                ))}
              </Stack>
            )}
            <Paper className="suggestion-box" radius="lg">
              <Group gap={10} wrap="nowrap">
                <ThemeIcon color="violet" variant="light" radius="xl"><Icon name="sparkle" size={16} /></ThemeIcon>
                <Box>
                  <Text fw={800} size="sm">{creationMode === 'group' ? 'Members only' : 'Open participation'}</Text>
                  <Text c="dimmed" size="xs">
                    {creationMode === 'group'
                      ? `Only members of the selected group can respond · POST /groups/{groupID}/${creationKind === 'poll' ? 'polls' : 'events'}/`
                      : `Anyone with access can respond · POST /${creationKind === 'poll' ? 'polls' : 'events'}/`}
                  </Text>
                </Box>
              </Group>
            </Paper>
            {creationError && <Text className="error-notice" role="alert">{creationError}</Text>}
            <Group justify="flex-end" mt="sm">
              <Button variant="subtle" color="dark" onClick={closeCreateModal}>Cancel</Button>
              <Button type="submit" className="create-button" loading={creationPending}>Create {creationKind}</Button>
            </Group>
          </Stack>
        </form>
      </Modal>

      <Modal
        opened={groupsOpened}
        onClose={closeGroupsModal}
        title={<span className="modal-title">Groups</span>}
        classNames={{ header: 'groups-modal-header' }}
        centered
        size="xl"
        padding={0}
        radius="xl"
      >
        <div className="groups-hero">
          <div className="groups-hero-icon"><Icon name="group" size={22} /></div>
          <Box>
            <Title order={3}>Manage groups</Title>
            <Text c="dimmed" size="sm" mt={3}>Create a group, join one with a code, or manage its members.</Text>
          </Box>
          <Group gap={7} className="groups-stats">
            <Badge variant="light" color="violet">{groups.filter((group) => group.role === 'owner').length} owned</Badge>
            <Badge variant="light" color="gray">{groups.filter((group) => group.role === 'member').length} joined</Badge>
          </Group>
        </div>

        <div className="group-workspace-content">
          {!managedGroup && (
            <SegmentedControl
              className="group-workspace-tabs"
              fullWidth
              value={groupAction}
              onChange={(value) => {
                setGroupAction(value as 'groups' | 'create' | 'join');
                setGroupError('');
                setActivationCode(null);
                setActivationCodeCopied(false);
              }}
              data={[
                { value: 'groups', label: 'My groups' },
                { value: 'create', label: 'Create group' },
                { value: 'join', label: 'Join with code' },
              ]}
            />
          )}

          {managedGroup ? (
            <section className="participant-view">
              <Group justify="space-between" align="center" mb="lg" wrap="nowrap">
                <Group gap={12} wrap="nowrap">
                  <ActionIcon variant="default" radius="md" size="lg" aria-label="Back to my groups" onClick={closeParticipantManager}>
                    <span className="back-arrow" aria-hidden="true">‹</span>
                  </ActionIcon>
                  <Box>
                    <Text className="group-form-kicker">MANAGE MEMBERS</Text>
                    <Title order={3}>{managedGroup.name}</Title>
                    <Text size="xs" c="dimmed" mt={2}>{managedGroup.participants.length + 1} people in this group</Text>
                  </Box>
                </Group>
                <Button
                  variant="light"
                  color="violet"
                  leftSection={<Icon name="plus" size={15} />}
                  loading={generatingGroupId === managedGroup.id}
                  onClick={() => generateActivationCode(managedGroup)}
                >
                  Invite people
                </Button>
              </Group>

              {activationCode && (
                <Paper className="activation-code" radius="lg" mb="md">
                  <Group justify="space-between" align="center" wrap="nowrap">
                    <Box>
                      <Text size="xs" fw={900}>INVITE CODE · {activationCode.groupName.toUpperCase()}</Text>
                      <Text component="code">{activationCode.code}</Text>
                      <Text size="xs" c="dimmed" mt={5}>Valid for 15 minutes · share privately</Text>
                    </Box>
                    <Button size="sm" variant="filled" color={activationCodeCopied ? 'green' : 'violet'} onClick={copyActivationCode}>
                      {activationCodeCopied ? 'Copied' : 'Copy code'}
                    </Button>
                  </Group>
                </Paper>
              )}

              <div className="participant-manager">
                <TextInput
                  value={participantQuery}
                  onChange={(event) => setParticipantQuery(event.currentTarget.value)}
                  placeholder="Search people…"
                  leftSection={<Icon name="search" size={17} />}
                  aria-label="Search group participants"
                />

                <div className="participant-toolbar">
                  <Checkbox
                    label={`Select all (${filteredParticipants.length})`}
                    checked={filteredParticipants.length > 0 && filteredParticipants.every((participant) => selectedParticipantIds.has(participant.id))}
                    indeterminate={filteredParticipants.some((participant) => selectedParticipantIds.has(participant.id))
                      && !filteredParticipants.every((participant) => selectedParticipantIds.has(participant.id))}
                    onChange={toggleAllVisibleParticipants}
                    disabled={filteredParticipants.length === 0}
                  />
                  <Text size="xs" c="dimmed">{managedGroup.participants.length + 1} people</Text>
                </div>

                <div className="participant-list">
                  <div className="participant-row is-owner">
                    <Avatar size={34} color="dark">AK</Avatar>
                    <Box className="participant-copy">
                      <Text size="sm" fw={850}>You</Text>
                      <Text size="xs" c="dimmed">Group owner</Text>
                    </Box>
                    <Badge size="xs" color="violet" variant="light">Owner</Badge>
                  </div>

                  {filteredParticipants.map((participant) => (
                    <label className={`participant-row ${selectedParticipantIds.has(participant.id) ? 'is-selected' : ''}`} key={participant.id}>
                      <Checkbox
                        checked={selectedParticipantIds.has(participant.id)}
                        onChange={() => toggleParticipant(participant.id)}
                        aria-label={`Select ${participant.name}`}
                      />
                      <Avatar size={34} color="violet" variant="light">{participant.initials}</Avatar>
                      <Box className="participant-copy">
                        <Text size="sm" fw={800}>{participant.name}</Text>
                        <Text size="xs" c="dimmed">{participant.email}</Text>
                      </Box>
                      <Text size="xs" c="dimmed">Member</Text>
                    </label>
                  ))}

                  {filteredParticipants.length === 0 && (
                    <Paper className="participant-empty" radius="lg">
                      <ThemeIcon color="gray" variant="light" radius="xl"><Icon name="user" size={16} /></ThemeIcon>
                      <Text size="sm" fw={800} mt={8}>{participantQuery ? 'No matching members' : 'No members yet'}</Text>
                      <Text size="xs" c="dimmed" mt={2}>{participantQuery ? 'Try another name or email.' : 'Generate an invite code to add people.'}</Text>
                    </Paper>
                  )}
                </div>

                {participantNotice && (
                  <Text className={/removed/i.test(participantNotice) ? 'success-notice' : 'error-notice'} role="status">
                    {participantNotice}
                  </Text>
                )}

                <div className="participant-actions">
                  <Text size="xs" c="dimmed">
                    {selectedParticipantIds.size === 0
                      ? 'Select members to manage'
                      : `${selectedParticipantIds.size} selected`}
                  </Text>
                  <Button
                    size="xs"
                    color="red"
                    variant="light"
                    loading={participantPending}
                    disabled={selectedParticipantIds.size === 0}
                    onClick={removeSelectedParticipants}
                  >
                    Remove from group
                  </Button>
                </div>
              </div>
            </section>
          ) : groupAction === 'groups' ? (
            <section className="group-list-view">
              <Group justify="space-between" align="flex-end" mb="lg">
                <Box>
                  <Text className="group-form-kicker">YOUR GROUPS</Text>
                  <Title order={3}>{groups.length} active {groups.length === 1 ? 'group' : 'groups'}</Title>
                  <Text size="sm" c="dimmed" mt={3}>Manage members or generate an invite code for groups you own.</Text>
                </Box>
                <Text size="xs" c="dimmed">Invite codes expire after 15 min</Text>
              </Group>

              {activationCode && (
                <Paper className="activation-code" radius="lg" mb="md">
                  <Group justify="space-between" align="center" wrap="nowrap">
                    <Box>
                      <Text size="xs" fw={900}>INVITE CODE · {activationCode.groupName.toUpperCase()}</Text>
                      <Text component="code">{activationCode.code}</Text>
                      <Text size="xs" c="dimmed" mt={5}>Valid for 15 minutes · share privately</Text>
                    </Box>
                    <Button size="sm" variant="filled" color={activationCodeCopied ? 'green' : 'violet'} onClick={copyActivationCode}>
                      {activationCodeCopied ? 'Copied' : 'Copy code'}
                    </Button>
                  </Group>
                </Paper>
              )}

              <Stack gap="sm" className="group-list">
                {groups.map((group) => (
                  <Paper className={`group-row ${activationCode?.groupId === group.id ? 'is-inviting' : ''}`} radius="lg" key={group.id}>
                    <span className="group-dot" style={{ background: group.color }} />
                    <Box className="group-row-copy">
                      <Text fw={850}>{group.name}</Text>
                      <Group gap={6} mt={4}>
                        <Badge size="xs" variant="light" color={group.role === 'owner' ? 'violet' : 'gray'}>{group.role}</Badge>
                        <Text size="xs" c="dimmed">{group.participants.length + 1} people · #{group.id}</Text>
                      </Group>
                    </Box>
                    {group.role === 'owner' ? (
                      <Group gap={6} wrap="nowrap" className="group-row-actions">
                        <Button size="xs" variant="default" leftSection={<Icon name="user" size={14} />} onClick={() => openParticipantManager(group)}>Manage members</Button>
                        <Button
                          size="xs"
                          variant={activationCode?.groupId === group.id ? 'filled' : 'light'}
                          color="violet"
                          leftSection={<Icon name="plus" size={14} />}
                          loading={generatingGroupId === group.id}
                          disabled={generatingGroupId !== null && generatingGroupId !== group.id}
                          onClick={() => generateActivationCode(group)}
                        >
                          {activationCode?.groupId === group.id ? 'Refresh code' : 'Invite'}
                        </Button>
                      </Group>
                    ) : (
                      <Badge size="sm" variant="outline" color="gray">Member</Badge>
                    )}
                  </Paper>
                ))}
              </Stack>
            </section>
          ) : (
            <section className="group-form-view">
              {groupAction === 'create' ? (
                <form onSubmit={createGroup} className="group-form">
                  <Text className="group-form-kicker">NEW GROUP</Text>
                  <Title order={3}>Create a group</Title>
                  <Text c="dimmed" size="sm" mt={4}>You will be the group owner.</Text>
                  <Stack gap="md" mt="xl">
                    <TextInput name="groupName" label="Group name" placeholder="e.g. Study group" required minLength={6} autoFocus />
                    <Paper className="api-id-field" radius="lg">
                      <Group gap={10} wrap="nowrap" align="flex-start">
                        <ThemeIcon color="gray" variant="light" radius="xl" size="sm"><Icon name="settings" size={13} /></ThemeIcon>
                        <Box>
                          <Text size="sm" fw={800}>API details</Text>
                          <Text size="xs" c="dimmed">Enter the ID of the matching API record.</Text>
                        </Box>
                      </Group>
                      <NumberInput
                        name="groupId"
                        label={session.kind === 'api' ? 'Group ID' : 'Demo group ID'}
                        placeholder={session.kind === 'api' ? 'Required, e.g. 24' : 'Optional · assigned automatically'}
                        min={1}
                        allowDecimal={false}
                        required={session.kind === 'api'}
                        mt="sm"
                      />
                    </Paper>
                    <Button type="submit" className="create-button" fullWidth loading={groupPending}>Create group</Button>
                  </Stack>
                </form>
              ) : (
                <form onSubmit={joinGroup} className="group-form">
                  <Text className="group-form-kicker">JOIN A GROUP</Text>
                  <Title order={3}>Use an activation code</Title>
                  <Text c="dimmed" size="sm" mt={4}>Ask the group owner for a code. It remains active for 15 minutes.</Text>
                  <Stack gap="md" mt="xl">
                    <TextInput
                      name="activationCode"
                      label="Activation code"
                      placeholder="Paste code"
                      leftSection={<Icon name="lock" size={16} />}
                      required
                      autoFocus
                    />
                    <Paper className="workspace-details" radius="lg">
                      <Text size="sm" fw={800}>Group details</Text>
                      <Text size="xs" c="dimmed" mt={2}>Enter the group name and ID used by the API.</Text>
                      <Group grow mt="sm">
                        <TextInput name="joinedGroupName" label="Group name" placeholder="Study group" required />
                        <NumberInput name="joinedGroupId" label="Group ID" placeholder="24" min={1} allowDecimal={false} required />
                      </Group>
                    </Paper>
                    <Button type="submit" className="create-button" fullWidth loading={groupPending}>Join group</Button>
                  </Stack>
                </form>
              )}
            </section>
          )}

          {groupError && <Text className="error-notice group-error" role="alert">{groupError}</Text>}
        </div>
      </Modal>
    </Box>
  );
}

const accessTokenStorageKey = 'fit-the-meet-access-token';

function storedSession(): Session | null {
  const accessToken = localStorage.getItem(accessTokenStorageKey) || sessionStorage.getItem(accessTokenStorageKey);
  return accessToken ? { kind: 'api', accessToken } : null;
}

function App() {
  const [authMode, setAuthMode] = useState<AuthMode>('login');
  const [session, setSession] = useState<Session | null>(storedSession);

  const login = async (credentials: LoginCredentials, remember: boolean) => {
    const { token } = await authApi.login(credentials);
    const storage = remember ? localStorage : sessionStorage;
    const otherStorage = remember ? sessionStorage : localStorage;
    storage.setItem(accessTokenStorageKey, token);
    otherStorage.removeItem(accessTokenStorageKey);
    setSession({ kind: 'api', accessToken: token });
  };

  const register = async (details: RegistrationDetails) => {
    await authApi.register(details);
  };

  const logout = () => {
    if (session?.kind === 'api') void authApi.logout(session.accessToken).catch(() => undefined);
    localStorage.removeItem(accessTokenStorageKey);
    sessionStorage.removeItem(accessTokenStorageKey);
    setSession(null);
    setAuthMode('login');
  };

  if (!session) {
    return (
      <AuthScreen
        mode={authMode}
        onModeChange={setAuthMode}
        onLogin={login}
        onRegister={register}
        onDemo={() => setSession({ kind: 'demo' })}
      />
    );
  }

  return <Dashboard session={session} onLogout={logout} />;
}

export default App;
