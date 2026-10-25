import { FormEvent, useMemo, useState } from 'react';
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
  LoginCredentials,
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

type AvailabilityDraft = AvailabilityTimeBlock & {
  key: number;
};

type Meeting = {
  id: number;
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
  options: TimeOption[];
};

function availabilityFromOption(option: TimeOption): AvailabilityDraft {
  const dateMatch = option.label.match(/(\d{1,2})\s(Oct|Nov)/);
  const month = dateMatch?.[2] === 'Nov' ? '11' : '10';
  const day = dateMatch?.[1].padStart(2, '0') || '26';
  const [startTime = '09:00', endTime = '10:00'] = option.detail.split('–');
  return {
    key: option.id,
    date: `2026-${month}-${day}`,
    start_time: startTime,
    end_time: endTime,
  };
}

function emptyAvailabilityBlock(offset = 0): AvailabilityDraft {
  return {
    key: Date.now() + offset,
    date: '2026-10-26',
    start_time: '09:00',
    end_time: '10:00',
  };
}

const initialMeetings: Meeting[] = [
  {
    id: 1,
    title: 'Q4 product direction',
    description: 'Choose the final slot for our roadmap review before the planning window closes.',
    group: 'Product team',
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
    options: [
      { id: 11, label: 'Monday, 26 Oct', detail: '10:00–10:50', votes: 8 },
      { id: 12, label: 'Tuesday, 27 Oct', detail: '14:00–14:50', votes: 6 },
      { id: 13, label: 'Wednesday, 28 Oct', detail: '09:00–09:50', votes: 5 },
    ],
  },
  {
    id: 2,
    title: 'Design critique · mobile flows',
    description: 'Weekly critique for onboarding, availability, and the new poll result screens.',
    group: 'Design circle',
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
    attendees: ['AR', 'AK', 'DB', 'IM'],
    options: [
      { id: 21, label: 'Tuesday, 27 Oct', detail: '11:30–12:10', votes: 6 },
      { id: 22, label: 'Tuesday, 27 Oct', detail: '15:00–15:40', votes: 4 },
    ],
  },
  {
    id: 3,
    title: 'Founders weekly',
    description: 'Cash flow, partnerships, launch risks, and the three decisions we need this week.',
    group: 'Core team',
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
    attendees: ['OP', 'MW', 'KS', 'AR'],
    options: [
      { id: 31, label: 'Thursday, 29 Oct', detail: '09:00–10:00', votes: 4 },
      { id: 32, label: 'Thursday, 29 Oct', detail: '16:00–17:00', votes: 3 },
      { id: 33, label: 'Friday, 30 Oct', detail: '10:00–11:00', votes: 2 },
    ],
  },
  {
    id: 4,
    title: 'API handoff & integration',
    description: 'Align on the poll endpoints, error states, and frontend integration milestones.',
    group: 'Engineering',
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
    options: [
      { id: 41, label: 'Wednesday, 28 Oct', detail: '13:00–13:45', votes: 9 },
    ],
  },
  {
    id: 5,
    title: 'Customer interview synthesis',
    description: 'Turn six recent interviews into concrete themes and next-step experiments.',
    group: 'Research crew',
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
    title: 'October team social',
    description: 'A low-key dinner. Vote for the slot first; we will pick the place together next.',
    group: 'Everyone',
    location: 'Warsaw · TBD',
    organizer: 'People Ops',
    dateLabel: '1 week left',
    duration: '2 hours',
    responses: 12,
    participants: 18,
    status: 'collecting',
    accent: '#ffb000',
    softAccent: '#fff8df',
    icon: 'calendar',
    attendees: ['PO', 'AR', 'MK', 'NS'],
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

const groupFilters = [
  { id: 'Product team', label: 'Product team', color: '#e9346f' },
  { id: 'Design circle', label: 'Design circle', color: '#6f45ed' },
  { id: 'Engineering', label: 'Engineering', color: '#078fa4' },
  { id: 'Core team', label: 'Core team', color: '#ff681d' },
];

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
          <Text className="eyebrow">Make time together</Text>
          <Title order={1}>One link.<br />One clear time.</Title>
          <Text className="auth-lede">Skip the message maze. Share a poll, see everyone’s availability, and settle on the best slot.</Text>
        </Box>

        <Paper className="auth-preview-card" radius="lg">
          <Group justify="space-between" align="flex-start" wrap="nowrap">
            <Box>
              <Badge className="auth-preview-badge">Product team</Badge>
              <Title order={3} mt={12}>Q4 product direction</Title>
              <Text c="dimmed" size="sm" mt={4}>Pick every time that works for you.</Text>
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
            <Text size="sm" fw={800}>Almost there <span aria-hidden="true">→</span></Text>
          </Group>
        </Paper>

        <Group className="auth-trust" gap={10} wrap="nowrap">
          <ThemeIcon color="dark" radius="xl" size={34}><Icon name="check" size={17} /></ThemeIcon>
          <Text size="sm"><strong>Designed for decisions,</strong><br />not calendar admin.</Text>
        </Group>
      </section>

      <section className="auth-form-pane">
        <Group gap={9} className="auth-mobile-brand">
          <div className="brand-mark"><Icon name="calendar" size={20} stroke={2.2} /></div>
          <div className="brand">fit<span>the</span>meet<i>.</i></div>
        </Group>

        <Paper className="auth-card" radius="xl">
          <Text className="eyebrow">{isLogin ? 'Welcome back' : 'Create your workspace'}</Text>
          <Title order={2} mt={5}>{isLogin ? 'Sign in to your events.' : 'Start finding better times.'}</Title>
          <Text c="dimmed" mt={9} size="sm">
            {isLogin ? 'Your events, groups, and decisions are waiting.' : 'Set up your account in under a minute.'}
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
                placeholder="you@company.com"
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
            Explore the demo workspace
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
  const [meetings, setMeetings] = useState(initialMeetings);
  const [query, setQuery] = useState('');
  const [activeFilter, setActiveFilter] = useState('all');
  const [sort, setSort] = useState('urgent');
  const [selectedMeeting, setSelectedMeeting] = useState<Meeting | null>(null);
  const [managedEventId, setManagedEventId] = useState<number | string>('');
  const [availabilityBlocks, setAvailabilityBlocks] = useState<AvailabilityDraft[]>([emptyAvailabilityBlock()]);
  const [eventActive, setEventActive] = useState(true);
  const [operationPending, setOperationPending] = useState(false);
  const [notice, setNotice] = useState('');
  const [dashboardNotice, setDashboardNotice] = useState('');
  const [creationMode, setCreationMode] = useState<'standalone' | 'group'>('standalone');
  const [creationPending, setCreationPending] = useState(false);
  const [creationError, setCreationError] = useState('');
  const [mobileFiltersOpened, mobileFilters] = useDisclosure(false);
  const [createOpened, createModal] = useDisclosure(false);

  const closeCreateModal = () => {
    setCreationError('');
    setCreationMode('standalone');
    createModal.close();
  };

  const filteredMeetings = useMemo(() => {
    const normalizedQuery = query.trim().toLowerCase();
    const result = meetings.filter((meeting) => {
      const matchesQuery = !normalizedQuery || [meeting.title, meeting.description, meeting.group, meeting.organizer]
        .join(' ')
        .toLowerCase()
        .includes(normalizedQuery);
      const matchesFilter = activeFilter === 'all'
        || (activeFilter === 'needs-response' && meeting.status === 'needs-response')
        || (activeFilter === 'ready' && meeting.status === 'ready')
        || (activeFilter === 'favorites' && meeting.favorite)
        || activeFilter === 'this-week'
        || meeting.group === activeFilter;
      return matchesQuery && matchesFilter;
    });
    return [...result].sort((a, b) => {
      if (sort === 'responses') return (b.responses / b.participants) - (a.responses / a.participants);
      if (sort === 'newest') return b.id - a.id;
      const priority: Record<MeetingStatus, number> = { 'needs-response': 0, collecting: 1, ready: 2, scheduled: 3 };
      return priority[a.status] - priority[b.status];
    });
  }, [activeFilter, meetings, query, sort]);

  const openMeeting = (meeting: Meeting) => {
    setSelectedMeeting(meeting);
    setManagedEventId(meeting.apiId === null ? '' : (meeting.apiId ?? meeting.id));
    setAvailabilityBlocks(meeting.options.slice(0, 2).map(availabilityFromOption));
    setEventActive(meeting.isActive ?? true);
    setNotice('');
  };

  const toggleFavorite = (id: number) => {
    setMeetings((current) => current.map((meeting) => meeting.id === id ? { ...meeting, favorite: !meeting.favorite } : meeting));
  };

  const parsedManagedEventId = typeof managedEventId === 'number' ? managedEventId : Number(managedEventId);

  const saveAvailability = async () => {
    if (!selectedMeeting || !Number.isInteger(parsedManagedEventId) || parsedManagedEventId <= 0) {
      setNotice('Enter a valid event ID before saving availability.');
      return;
    }

    if (!eventActive) {
      setNotice('Activate the event before submitting availability.');
      return;
    }

    if (availabilityBlocks.length === 0) {
      setNotice('Add at least one availability block.');
      return;
    }

    const invalidBlock = availabilityBlocks.find((block) => (
      !block.date || !block.start_time || !block.end_time || block.start_time >= block.end_time
    ));
    if (invalidBlock) {
      setNotice('Every block needs a date and a start time before its end time.');
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
      setNotice(session.kind === 'demo' ? 'Availability saved in demo mode.' : 'Availability saved to your API.');
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

    setOperationPending(true);
    setNotice('');
    try {
      if (session.kind === 'api') await eventApi.toggleActive(parsedManagedEventId, session.accessToken);
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
    const title = String(form.get('title') || 'Untitled event');
    const durationMinutes = Number(form.get('duration') || 30);
    const groupId = creationMode === 'group' ? Number(form.get('groupId')) : undefined;
    if (creationMode === 'group' && (!Number.isInteger(groupId) || Number(groupId) <= 0)) {
      setCreationError('Enter a valid group ID.');
      return;
    }

    setCreationPending(true);
    setCreationError('');
    try {
      if (session.kind === 'api') {
        await eventApi.create({
          name: title,
          description: String(form.get('description') || ''),
          duration_minutes: durationMinutes,
        }, session.accessToken, groupId);
      }

      const nextId = Math.max(...meetings.map((meeting) => meeting.id)) + 1;
      const group = groupId ? `Group #${groupId}` : 'Personal';
      const next: Meeting = {
        id: nextId,
        title,
        description: String(form.get('description') || 'A new event waiting for everyone’s availability.'),
        group,
        location: 'Remote',
        organizer: 'You',
        dateLabel: 'Just created',
        duration: `${durationMinutes} min`,
        responses: 0,
        participants: 5,
        status: 'collecting',
        accent: '#6f45ed',
        softAccent: '#f3efff',
        icon: 'calendar',
        attendees: ['YOU'],
        isActive: true,
        apiId: session.kind === 'demo' ? nextId : null,
        options: [
          { id: Date.now(), label: 'Monday, 2 Nov', detail: '10:00–10:30', votes: 1 },
          { id: Date.now() + 1, label: 'Tuesday, 3 Nov', detail: '14:00–14:30', votes: 1 },
        ],
      };
      setMeetings((current) => [next, ...current]);
      setActiveFilter('all');
      setDashboardNotice(session.kind === 'demo'
        ? 'Event created in demo mode.'
        : 'Event created through your API. The current create response does not include its new ID.');
      closeCreateModal();
    } catch (requestError) {
      setCreationError(errorMessage(requestError));
    } finally {
      setCreationPending(false);
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
        <Text className="sidebar-kicker">Groups</Text>
        <Stack gap={5} mt={12}>
          {groupFilters.map((item) => (
            <UnstyledButton
              key={item.id}
              className={`sidebar-item ${activeFilter === item.id ? 'is-active' : ''}`}
              onClick={() => { setActiveFilter(item.id); mobileFilters.close(); }}
            >
              <span className="group-dot" style={{ background: item.color }} />
              <span>{item.label}</span>
              <span className="sidebar-count">{countFor(item.id)}</span>
            </UnstyledButton>
          ))}
        </Stack>
      </Box>
      <Paper className="sidebar-tip" radius="xl">
        <ThemeIcon color="dark" variant="filled" radius="xl" size={34}><Icon name="sparkle" size={17} /></ThemeIcon>
        <Text fw={800} mt={12}>Fewer messages. Faster plans.</Text>
        <Text size="sm" c="dimmed" mt={5}>Invite people once and let availability do the work.</Text>
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
            placeholder="Search events, groups, people…"
            leftSection={<Icon name="search" size={21} />}
            variant="unstyled"
            aria-label="Search events"
          />
          <Divider orientation="vertical" />
          <Button className="workspace-button" variant="subtle" color="dark" leftSection={<Icon name="group" size={19} />}>
            Event workspace
          </Button>
          <Divider orientation="vertical" />
          <ActionIcon variant="subtle" color="gray" radius="xl" size="lg" aria-label="Workspace settings">
            <Icon name="settings" size={21} />
          </ActionIcon>
        </Paper>

        <Group gap={10} wrap="nowrap">
          {session.kind === 'demo' && <Badge className="mode-badge">Demo</Badge>}
          <Button className="create-button" leftSection={<Icon name="plus" size={19} />} onClick={createModal.open}>New event</Button>
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
              <Title order={1}>Plan events without the chase.</Title>
              <Text c="dimmed" mt={6}>
                You have <strong>{countFor('needs-response')} {countFor('needs-response') === 1 ? 'event' : 'events'}</strong> waiting for your availability.
              </Text>
            </Box>
            <Select
              className="sort-select"
              value={sort}
              onChange={(value) => setSort(value || 'urgent')}
              data={[
                { value: 'urgent', label: 'Most urgent' },
                { value: 'newest', label: 'Newest first' },
                { value: 'responses', label: 'Most responses' },
              ]}
              allowDeselect={false}
              leftSection={<Icon name="filter" size={17} />}
              aria-label="Sort events"
            />
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
              <Text size="sm"><strong>Connected to your Go API.</strong> Your backend has mutation routes but no event-list route yet, so the cards below remain workspace previews. Enter the real event ID in the event drawer before updating it.</Text>
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
              <Button variant="light" color="dark" mt={18} onClick={() => { setQuery(''); setActiveFilter('all'); }}>Clear filters</Button>
            </Paper>
          )}
        </section>
      </main>

      <Drawer opened={mobileFiltersOpened} onClose={mobileFilters.close} title={<span className="drawer-title">Browse events</span>} size="xs" padding="lg">
        <SidebarContent />
      </Drawer>

      <Drawer
        opened={Boolean(selectedMeeting)}
        onClose={() => setSelectedMeeting(null)}
        position="right"
        size={480}
        padding="xl"
        title={<span className="drawer-title">Event workspace</span>}
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
            </Group>

            <Paper className="event-control-panel" radius="xl">
              <Group justify="space-between" align="flex-start" wrap="nowrap">
                <Box>
                  <Text fw={850}>API event</Text>
                  <Text size="xs" c="dimmed" mt={2}>Used by status and availability actions.</Text>
                </Box>
                <Badge color={eventActive ? 'green' : 'gray'} variant="light">{eventActive ? 'Active' : 'Inactive'}</Badge>
              </Group>
              <NumberInput
                label="Event ID"
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

            <Box>
              <Group justify="space-between" align="flex-start">
                <Box>
                  <Title order={3}>Your availability</Title>
                  <Text size="sm" c="dimmed" mt={4}>Add one or more date and time ranges.</Text>
                </Box>
                <Button
                  size="xs"
                  variant="light"
                  color="dark"
                  leftSection={<Icon name="plus" size={15} />}
                  onClick={() => setAvailabilityBlocks((current) => [...current, emptyAvailabilityBlock(current.length)])}
                >
                  Add block
                </Button>
              </Group>
              <Stack gap={12} mt={15}>
                {availabilityBlocks.map((block, index) => (
                  <Paper className="availability-block" radius="lg" key={block.key}>
                    <Group justify="space-between" mb={10}>
                      <Text size="xs" fw={850} c="dimmed">TIME BLOCK {index + 1}</Text>
                      <Button
                        size="compact-xs"
                        variant="subtle"
                        color="red"
                        disabled={availabilityBlocks.length === 1}
                        onClick={() => setAvailabilityBlocks((current) => current.filter((item) => item.key !== block.key))}
                      >
                        Remove
                      </Button>
                    </Group>
                    <TextInput
                      type="date"
                      label="Date"
                      value={block.date}
                      onChange={(event) => setAvailabilityBlocks((current) => current.map((item) => item.key === block.key ? { ...item, date: event.currentTarget.value } : item))}
                    />
                    <Group grow mt="sm">
                      <TextInput
                        type="time"
                        label="Start"
                        value={block.start_time}
                        onChange={(event) => setAvailabilityBlocks((current) => current.map((item) => item.key === block.key ? { ...item, start_time: event.currentTarget.value } : item))}
                      />
                      <TextInput
                        type="time"
                        label="End"
                        value={block.end_time}
                        onChange={(event) => setAvailabilityBlocks((current) => current.map((item) => item.key === block.key ? { ...item, end_time: event.currentTarget.value } : item))}
                      />
                    </Group>
                  </Paper>
                ))}
              </Stack>
            </Box>

            {notice && <Text role="status" className={/(saved|now active|now inactive)/i.test(notice) ? 'success-notice' : 'error-notice'}>{notice}</Text>}
            <Button className="save-vote-button" size="lg" loading={operationPending} disabled={!eventActive} onClick={saveAvailability}>Save availability</Button>
          </Stack>
        )}
      </Drawer>

      <Modal opened={createOpened} onClose={closeCreateModal} title={<span className="modal-title">Create an event</span>} centered size="lg" padding="xl" radius="xl">
        <form onSubmit={createMeeting}>
          <Stack gap="md">
            <Text c="dimmed" size="sm">Choose the standalone or group endpoint, then send the exact payload expected by your Go API.</Text>
            <SegmentedControl
              fullWidth
              value={creationMode}
              onChange={(value) => { setCreationMode(value as 'standalone' | 'group'); setCreationError(''); }}
              data={[
                { value: 'standalone', label: 'Standalone event' },
                { value: 'group', label: 'Group event' },
              ]}
            />
            <TextInput name="title" label="Event name" placeholder="e.g. November planning" required minLength={6} />
            <Textarea name="description" label="What is it about?" placeholder="Add a short agenda or context" minRows={3} />
            {creationMode === 'group' && (
              <NumberInput name="groupId" label="Group ID" placeholder="e.g. 12" min={1} allowDecimal={false} required />
            )}
            <NumberInput name="duration" label="Duration (minutes)" defaultValue={30} min={1} max={1440} step={15} required />
            <Paper className="suggestion-box" radius="lg">
              <Group gap={10} wrap="nowrap">
                <ThemeIcon color="violet" variant="light" radius="xl"><Icon name="sparkle" size={16} /></ThemeIcon>
                <Box><Text fw={800} size="sm">Exact endpoint</Text><Text c="dimmed" size="xs">{creationMode === 'group' ? 'POST /groups/{groupID}/events/' : 'POST /events/'}</Text></Box>
              </Group>
            </Paper>
            {creationError && <Text className="error-notice" role="alert">{creationError}</Text>}
            <Group justify="flex-end" mt="sm">
              <Button variant="subtle" color="dark" onClick={closeCreateModal}>Cancel</Button>
              <Button type="submit" className="create-button" loading={creationPending}>Create event</Button>
            </Group>
          </Stack>
        </form>
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
