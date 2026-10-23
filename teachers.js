(() => {
  const search = document.getElementById('teacherSearch');
  if (!search) return;
  const profiles = [...document.querySelectorAll('.head-teacher, .teacher-card')];
  const searchTexts = profiles.map(profile => {
    const current = profile.textContent;
    const terms = Object.entries(typeof SCHOOL_BANGLA === 'undefined' ? {} : SCHOOL_BANGLA)
      .filter(([en,bn]) => current.includes(en) || current.includes(bn)).flat();
    return `${current} ${terms.join(' ')}`.toLocaleLowerCase();
  });
  const count = document.getElementById('teacherCount');
  const empty = document.getElementById('noTeachers');
  function update() {
    const query = search.value.trim().toLocaleLowerCase();
    let visible = 0;
    profiles.forEach((profile,index) => {
      profile.hidden = !searchTexts[index].includes(query);
      if (!profile.hidden) visible++;
    });
    count.textContent = document.documentElement.lang === 'bn'
      ? `${visible.toLocaleString('bn-BD')}টি নমুনা পরিচিতি দেখানো হচ্ছে`
      : `${visible} preview ${visible === 1 ? 'profile' : 'profiles'} displayed`;
    empty.hidden = visible !== 0;
  }
  search.addEventListener('input',update);
  document.addEventListener('schoollanguagechange',update);
  update();
})();
