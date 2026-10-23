(() => {
 const tabs=[...document.querySelectorAll('[role="tab"]')];
 const panels=[...document.querySelectorAll('[role="tabpanel"]')];
 const bn=()=>document.documentElement.lang==='bn';
 const tr=(en,bangla)=>bn()?bangla:en;
 function activate(id,focus=false){
  if(!panels.some(p=>p.id===id)) id='curriculum';
  tabs.forEach(tab=>{const active=tab.getAttribute('aria-controls')===id;tab.setAttribute('aria-selected',String(active));tab.tabIndex=active?0:-1;if(active&&focus)tab.focus();});
  panels.forEach(panel=>panel.hidden=panel.id!==id);
 }
 tabs.forEach((tab,index)=>{
  tab.addEventListener('click',()=>{const id=tab.getAttribute('aria-controls');activate(id);history.replaceState(null,'',`#${id}`);});
  tab.addEventListener('keydown',event=>{
   let next;
   if(event.key==='ArrowRight')next=(index+1)%tabs.length;
   if(event.key==='ArrowLeft')next=(index+tabs.length-1)%tabs.length;
   if(event.key==='Home')next=0;
   if(event.key==='End')next=tabs.length-1;
   if(next!==undefined){event.preventDefault();activate(tabs[next].getAttribute('aria-controls'),true);}
  });
 });
 window.addEventListener('hashchange',()=>activate(location.hash.slice(1)));
 activate(location.hash.slice(1));
 const classSelect=document.getElementById('curriculumClass');
 const group=document.getElementById('curriculumGroup');
 let category='all',searched=false;
 const resultForm=document.getElementById('resultFilters');
 function render(){
  document.getElementById('curriculumGroupField').hidden=Number(classSelect.value)<9;
  const selected=classSelect.value;
  document.querySelector('#curriculum .resource-grid').hidden=!selected;
  document.getElementById('classResourceStatus').textContent=selected
   ? tr(`Resources for Class ${selected}${Number(selected)>=9?' · '+group.selectedOptions[0].textContent:''}`,`শ্রেণি ${Number(selected).toLocaleString('bn-BD')}${Number(selected)>=9?' · '+group.selectedOptions[0].textContent:''}-এর তথ্য`)
   :tr('Choose a class to view available resources.','তথ্য দেখতে শ্রেণি নির্বাচন করুন।');
  document.getElementById('resultStatus').textContent=searched
   ?tr(`No result sheet is published for ${document.getElementById('resultYear').value}, Class ${document.getElementById('resultClass').value}, ${document.getElementById('resultExam').selectedOptions[0].textContent}.`,`${Number(document.getElementById('resultYear').value).toLocaleString('bn-BD',{useGrouping:false})} সালের শ্রেণি ${Number(document.getElementById('resultClass').value).toLocaleString('bn-BD')}, ${document.getElementById('resultExam').selectedOptions[0].textContent} পরীক্ষার ফলাফল এখনও প্রকাশিত হয়নি।`)
   :tr('Select your class and examination, then choose Find results.','শ্রেণি ও পরীক্ষা নির্বাচন করে ফলাফল খুঁজুন বাটনে চাপুন।');
  const labels={all:['dates','তারিখ'],exams:['exam dates','পরীক্ষার তারিখ'],holidays:['holiday dates','ছুটির তারিখ'],events:['event dates','অনুষ্ঠানের তারিখ']};
  document.getElementById('calendarStatus').textContent=tr(`No upcoming ${labels[category][0]} have been published. Check the notice board or contact the school office.`,`আসন্ন ${labels[category][1]} এখনও প্রকাশিত হয়নি। বিজ্ঞপ্তি বোর্ড দেখুন বা স্কুল অফিসে যোগাযোগ করুন।`);
 }
 classSelect.addEventListener('change',render);group.addEventListener('change',render);
 resultForm.addEventListener('submit',event=>{event.preventDefault();searched=true;render();});
 resultForm.addEventListener('change',()=>{searched=false;render();});
 document.querySelectorAll('[data-category]').forEach(button=>button.addEventListener('click',()=>{category=button.dataset.category;document.querySelectorAll('[data-category]').forEach(b=>b.setAttribute('aria-pressed',String(b===button)));render();}));
 document.addEventListener('schoollanguagechange',render);render();
})();
