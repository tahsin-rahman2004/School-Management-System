document.querySelectorAll('input[type="password"]').forEach(input => {
  const wrapper = document.createElement('div');
  wrapper.className = 'password-field';
  input.before(wrapper);
  wrapper.append(input);

  const toggle = document.createElement('button');
  toggle.type = 'button';
  toggle.className = 'password-toggle';
  toggle.setAttribute('aria-controls', input.id);
  toggle.setAttribute('aria-pressed', 'false');
  toggle.innerHTML = '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" aria-hidden="true"><path d="M2 12s3.5-7 10-7 10 7 10 7-3.5 7-10 7S2 12 2 12Z"/><circle cx="12" cy="12" r="3"/><path class="eye-slash" d="m3 3 18 18"/></svg>';

  function updateLabel() {
    const visible = input.type === 'text';
    const bangla = document.documentElement.lang === 'bn';
    const label = bangla
      ? (visible ? 'পাসওয়ার্ড লুকান' : 'পাসওয়ার্ড দেখুন')
      : (visible ? 'Hide password' : 'Show password');
    toggle.setAttribute('aria-label', label);
    toggle.title = label;
    toggle.setAttribute('aria-pressed', String(visible));
  }

  toggle.addEventListener('click', () => {
    input.type = input.type === 'password' ? 'text' : 'password';
    updateLabel();
  });
  document.addEventListener('schoollanguagechange', updateLabel);
  wrapper.append(toggle);
  updateLabel();
});

const attendanceSubject = document.getElementById('attendance-subject');
const attendanceMessage = document.getElementById('attendance-message');
if (attendanceSubject && attendanceMessage) {
  function updateAttendanceMessage() {
    const bangla = document.documentElement.lang === 'bn';
    const locale = bangla ? 'bn-BD' : 'en-GB';
    const subject = attendanceSubject.selectedOptions[0].textContent;
    const history = document.getElementById('attendance-history');
    const rows = [...document.querySelectorAll('#attendance-rows tr[data-subject]')];
    let count = 0;
    let attended = 0;
    let held = 0;
    rows.sort((a, b) => b.dataset.date.localeCompare(a.dataset.date)).forEach(row => {
      const selected = Boolean(attendanceSubject.value) && row.dataset.subject === attendanceSubject.value;
      row.hidden = !selected;
      if (selected) count++;
      const date = new Date(`${row.dataset.date}T12:00:00`);
      if (!Number.isNaN(date.getTime())) {
        row.cells[0].textContent = date.toLocaleDateString(locale, { day: 'numeric', month: 'short', year: 'numeric' });
        row.cells[1].textContent = date.toLocaleDateString(locale, { weekday: 'long' });
      }
      const status = row.dataset.status;
      if (selected) {
        if (['present', 'late', 'absent'].includes(status)) held++;
        if (['present', 'late'].includes(status)) attended++;
      }
      const labels = { present: ['Present', 'উপস্থিত'], absent: ['Absent', 'অনুপস্থিত'], late: ['Late', 'বিলম্বে উপস্থিত'], excused: ['Excused', 'অনুমোদিত ছুটি'] };
      row.cells[2].textContent = labels[status]?.[bangla ? 1 : 0] || (bangla ? 'নথিভুক্ত হয়নি' : 'Not recorded');
      row.cells[2].className = Object.hasOwn(labels, status) ? `attendance-${status}` : '';
      row.parentElement.insertBefore(row, document.getElementById('attendance-no-records'));
    });
    const summary = document.getElementById('attendance-summary');
    summary.hidden = !attendanceSubject.value;
    document.getElementById('attendance-percentage').textContent = held
      ? new Intl.NumberFormat(locale, { style: 'percent', maximumFractionDigits: 1 }).format(attended / held)
      : '—';
    document.getElementById('attendance-total').textContent = `${attended.toLocaleString(locale)} / ${held.toLocaleString(locale)}`;
    history.hidden = !attendanceSubject.value;
    document.getElementById('attendance-no-records').hidden = count > 0;
    attendanceMessage.textContent = attendanceSubject.value
      ? (bangla ? `${subject} — দৈনিক উপস্থিতির তালিকা` : `${subject} — daily attendance history`)
      : (bangla ? 'কোনো বিষয় নির্বাচন করা হয়নি।' : 'No subject selected.');
  }
  attendanceSubject.addEventListener('change', updateAttendanceMessage);
  document.addEventListener('schoollanguagechange', updateAttendanceMessage);
  updateAttendanceMessage();
}

const studentClass = document.getElementById('class');
const studentGroup = document.getElementById('student-group');
const groupField = document.getElementById('student-group-field');
if (studentClass && studentGroup && groupField) {
  function updateStudentGroup() {
    const needsGroup = ['9', '10'].includes(studentClass.value);
    groupField.hidden = !needsGroup;
    studentGroup.required = needsGroup;
    studentGroup.disabled = !needsGroup;
    if (!needsGroup) studentGroup.value = '';
  }
  studentClass.addEventListener('change', updateStudentGroup);
  updateStudentGroup();
}
