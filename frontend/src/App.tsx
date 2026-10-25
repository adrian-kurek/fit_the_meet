import { FormEvent, useMemo, useState } from 'react';
import {
  ActionIcon,
  Avatar,
  Badge,
  Box,
  Burger,
  Button,
  Divider,
  Drawer,
  Group,
  Modal,
  NumberInput,
  Paper,
  Progress,
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

type IconName =
  | 'calendar'
  | 'check'
  | 'clock'
  | 'filter'
  | 'group'
  | 'heart'
  | 'home'
  | 'location'
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
  options: TimeOption[];
};

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
  { id: 'all', label: 'All meetings', icon: 'home' as IconName, color: '#252525' },
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
    location: <><path d="M20 10c0 5-8 11-8 11S4 15 4 10a8 8 0 1 1 16 0Z"/><circle cx="12" cy="10" r="2.5"/></>,
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

function App() {
  const [meetings, setMeetings] = useState(initialMeetings);
  const [query, setQuery] = useState('');
  const [activeFilter, setActiveFilter] = useState('all');
  const [sort, setSort] = useState('urgent');
  const [selectedMeeting, setSelectedMeeting] = useState<Meeting | null>(null);
  const [selectedOptions, setSelectedOptions] = useState<Set<number>>(new Set());
  const [notice, setNotice] = useState('');
  const [mobileFiltersOpened, mobileFilters] = useDisclosure(false);
  const [createOpened, createModal] = useDisclosure(false);

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
    setSelectedOptions(new Set());
    setNotice('');
  };

  const toggleFavorite = (id: number) => {
    setMeetings((current) => current.map((meeting) => meeting.id === id ? { ...meeting, favorite: !meeting.favorite } : meeting));
  };

  const saveVote = () => {
    if (!selectedMeeting || selectedOptions.size === 0) {
      setNotice('Select at least one time that works for you.');
      return;
    }
    setMeetings((current) => current.map((meeting) => meeting.id === selectedMeeting.id
      ? { ...meeting, status: 'collecting', responses: Math.min(meeting.responses + 1, meeting.participants) }
      : meeting));
    setNotice('Availability saved — thanks!');
  };

  const createMeeting = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    const form = new FormData(event.currentTarget);
    const title = String(form.get('title') || 'Untitled meeting');
    const group = String(form.get('group') || 'Personal');
    const duration = `${String(form.get('duration') || 30)} min`;
    const next: Meeting = {
      id: Math.max(...meetings.map((meeting) => meeting.id)) + 1,
      title,
      description: String(form.get('description') || 'A new meeting waiting for everyone’s availability.'),
      group,
      location: 'Remote',
      organizer: 'You',
      dateLabel: 'Just created',
      duration,
      responses: 1,
      participants: 5,
      status: 'collecting',
      accent: '#6f45ed',
      softAccent: '#f3efff',
      icon: 'calendar',
      attendees: ['YOU'],
      options: [
        { id: Date.now(), label: 'Monday, 2 Nov', detail: '10:00–10:30', votes: 1 },
        { id: Date.now() + 1, label: 'Tuesday, 3 Nov', detail: '14:00–14:30', votes: 1 },
      ],
    };
    setMeetings((current) => [next, ...current]);
    setActiveFilter('all');
    createModal.close();
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
          <Burger opened={mobileFiltersOpened} onClick={mobileFilters.toggle} hiddenFrom="md" size="sm" aria-label="Open meeting filters" />
          <div className="brand-mark"><Icon name="calendar" size={22} stroke={2.2} /></div>
          <div className="brand">fit<span>the</span>meet<i>.</i></div>
        </Group>

        <Paper className="search-shell" radius="xl">
          <TextInput
            value={query}
            onChange={(event) => setQuery(event.currentTarget.value)}
            placeholder="Search meetings, groups, people…"
            leftSection={<Icon name="search" size={21} />}
            variant="unstyled"
            aria-label="Search meetings"
          />
          <Divider orientation="vertical" />
          <Button className="workspace-button" variant="subtle" color="dark" leftSection={<Icon name="group" size={19} />}>
            My workspace
          </Button>
          <Divider orientation="vertical" />
          <ActionIcon variant="subtle" color="gray" radius="xl" size="lg" aria-label="Workspace settings">
            <Icon name="settings" size={21} />
          </ActionIcon>
        </Paper>

        <Group gap={10} wrap="nowrap">
          <Button className="create-button" leftSection={<Icon name="plus" size={19} />} onClick={createModal.open}>New meeting</Button>
          <Avatar color="dark" radius="xl">AK</Avatar>
        </Group>
      </header>

      <main className="workspace-layout">
        <aside className="desktop-sidebar"><SidebarContent /></aside>

        <section className="content-column">
          <Group justify="space-between" align="flex-end" className="content-heading">
            <Box>
              <Text className="eyebrow">Sunday, 25 October</Text>
              <Title order={1}>Find the time that fits.</Title>
              <Text c="dimmed" mt={6}>You have <strong>{countFor('needs-response')} meetings</strong> waiting for your response.</Text>
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
              aria-label="Sort meetings"
            />
          </Group>

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
              <Title order={3} mt={16}>No meetings found</Title>
              <Text c="dimmed" mt={5}>Try a different search or filter.</Text>
              <Button variant="light" color="dark" mt={18} onClick={() => { setQuery(''); setActiveFilter('all'); }}>Clear filters</Button>
            </Paper>
          )}
        </section>
      </main>

      <Drawer opened={mobileFiltersOpened} onClose={mobileFilters.close} title={<span className="drawer-title">Browse meetings</span>} size="xs" padding="lg">
        <SidebarContent />
      </Drawer>

      <Drawer
        opened={Boolean(selectedMeeting)}
        onClose={() => setSelectedMeeting(null)}
        position="right"
        size={480}
        padding="xl"
        title={<span className="drawer-title">Meeting poll</span>}
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

            <Paper className="response-panel" radius="xl">
              <Group justify="space-between">
                <Text fw={800}>Response progress</Text>
                <Text fw={800}>{Math.round(selectedMeeting.responses / selectedMeeting.participants * 100)}%</Text>
              </Group>
              <Progress value={selectedMeeting.responses / selectedMeeting.participants * 100} color="dark" radius="xl" size="md" mt={12} />
              <Group justify="space-between" mt={14}>
                <Avatar.Group spacing="sm">
                  {selectedMeeting.attendees.map((attendee, index) => <Avatar key={attendee} color={['violet', 'pink', 'cyan', 'orange'][index]} size="sm">{attendee}</Avatar>)}
                </Avatar.Group>
                <Text size="sm" c="dimmed">{selectedMeeting.responses} of {selectedMeeting.participants} replied</Text>
              </Group>
            </Paper>

            <Box>
              <Title order={3}>When can you make it?</Title>
              <Text size="sm" c="dimmed" mt={4}>Select every option that works.</Text>
              <Stack gap={11} mt={15}>
                {selectedMeeting.options.map((option) => {
                  const selected = selectedOptions.has(option.id);
                  return (
                    <UnstyledButton
                      key={option.id}
                      className={`time-option ${selected ? 'is-selected' : ''}`}
                      onClick={() => setSelectedOptions((current) => {
                        const next = new Set(current);
                        if (next.has(option.id)) next.delete(option.id); else next.add(option.id);
                        return next;
                      })}
                    >
                      <span className="option-check">{selected && <Icon name="check" size={17} stroke={2.4} />}</span>
                      <span className="option-copy"><strong>{option.label}</strong><small>{option.detail}</small></span>
                      <span className="option-votes">{option.votes} free</span>
                    </UnstyledButton>
                  );
                })}
              </Stack>
            </Box>

            {notice && <Text className={notice.startsWith('Availability') ? 'success-notice' : 'error-notice'}>{notice}</Text>}
            <Button className="save-vote-button" size="lg" onClick={saveVote}>Save my availability</Button>
          </Stack>
        )}
      </Drawer>

      <Modal opened={createOpened} onClose={createModal.close} title={<span className="modal-title">Create a meeting poll</span>} centered size="lg" padding="xl" radius="xl">
        <form onSubmit={createMeeting}>
          <Stack gap="md">
            <Text c="dimmed" size="sm">Start with the essentials. You can fine-tune time options before inviting anyone.</Text>
            <TextInput name="title" label="Meeting title" placeholder="e.g. November planning" required />
            <Textarea name="description" label="What is it about?" placeholder="Add a short agenda or context" minRows={3} />
            <Group grow align="flex-start">
              <Select name="group" label="Group" defaultValue="Product team" data={['Product team', 'Design circle', 'Engineering', 'Core team', 'Personal']} allowDeselect={false} />
              <NumberInput name="duration" label="Duration (minutes)" defaultValue={30} min={15} max={240} step={15} />
            </Group>
            <Paper className="suggestion-box" radius="lg">
              <Group gap={10} wrap="nowrap">
                <ThemeIcon color="violet" variant="light" radius="xl"><Icon name="sparkle" size={16} /></ThemeIcon>
                <Box><Text fw={800} size="sm">Smart start</Text><Text c="dimmed" size="xs">We’ll suggest two slots based on your working hours.</Text></Box>
              </Group>
            </Paper>
            <Group justify="flex-end" mt="sm">
              <Button variant="subtle" color="dark" onClick={createModal.close}>Cancel</Button>
              <Button type="submit" className="create-button">Create poll</Button>
            </Group>
          </Stack>
        </form>
      </Modal>
    </Box>
  );
}

export default App;
