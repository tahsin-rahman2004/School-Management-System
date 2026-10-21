(() => {
  const search = document.getElementById('teacherSearch');
  if (!search) return;
  const profiles = [...document.querySelectorAll('.head-teacher, .teacher-card')];
  const count = document.getElementById('teacherCount');
  const empty = document.getElementById('noTeachers');
  search.addEventListener('input', () => {
    const query = search.value.trim().toLocaleLowerCase();
    let visible = 0;
    profiles.forEach(profile => {
      profile.hidden = !profile.textContent.toLocaleLowerCase().includes(query);
      if (!profile.hidden) visible++;
    });
    count.textContent = `${visible} preview ${visible === 1 ? 'profile' : 'profiles'} displayed`;
    empty.hidden = visible !== 0;
  });
})();
