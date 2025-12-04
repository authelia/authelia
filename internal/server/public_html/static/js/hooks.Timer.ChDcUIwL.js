import{s as R,t as D,r as c,v as q,aT as M,j as p,w as y,x as T,F as i,y as w,E as C,H as b,aU as k,aV as L}from"./index.CEiXDjRd.js";function N(e){return R("MuiLinearProgress",e)}D("MuiLinearProgress",["root","colorPrimary","colorSecondary","determinate","indeterminate","buffer","query","dashed","dashedColorPrimary","dashedColorSecondary","bar","bar1","bar2","barColorPrimary","barColorSecondary","bar1Indeterminate","bar1Determinate","bar1Buffer","bar2Indeterminate","bar2Buffer"]);const h=4,P=L`
  0% {
    left: -35%;
    right: 100%;
  }

  60% {
    left: 100%;
    right: -90%;
  }

  100% {
    left: 100%;
    right: -90%;
  }
`,O=typeof P!="string"?k`
        animation: ${P} 2.1s cubic-bezier(0.65, 0.815, 0.735, 0.395) infinite;
      `:null,x=L`
  0% {
    left: -200%;
    right: 100%;
  }

  60% {
    left: 107%;
    right: -8%;
  }

  100% {
    left: 107%;
    right: -8%;
  }
`,S=typeof x!="string"?k`
        animation: ${x} 2.1s cubic-bezier(0.165, 0.84, 0.44, 1) 1.15s infinite;
      `:null,$=L`
  0% {
    opacity: 1;
    background-position: 0 -23px;
  }

  60% {
    opacity: 0;
    background-position: 0 -23px;
  }

  100% {
    opacity: 1;
    background-position: -200px -23px;
  }
`,z=typeof $!="string"?k`
        animation: ${$} 3s infinite linear;
      `:null,A=e=>{const{classes:r,variant:a,color:t}=e,u={root:["root",`color${i(t)}`,a],dashed:["dashed",`dashedColor${i(t)}`],bar1:["bar","bar1",`barColor${i(t)}`,(a==="indeterminate"||a==="query")&&"bar1Indeterminate",a==="determinate"&&"bar1Determinate",a==="buffer"&&"bar1Buffer"],bar2:["bar","bar2",a!=="buffer"&&`barColor${i(t)}`,a==="buffer"&&`color${i(t)}`,(a==="indeterminate"||a==="query")&&"bar2Indeterminate",a==="buffer"&&"bar2Buffer"]};return w(u,N,r)},I=(e,r)=>e.vars?e.vars.palette.LinearProgress[`${r}Bg`]:e.palette.mode==="light"?e.lighten(e.palette[r].main,.62):e.darken(e.palette[r].main,.5),U=y("span",{name:"MuiLinearProgress",slot:"Root",overridesResolver:(e,r)=>{const{ownerState:a}=e;return[r.root,r[`color${i(a.color)}`],r[a.variant]]}})(C(({theme:e})=>({position:"relative",overflow:"hidden",display:"block",height:4,zIndex:0,"@media print":{colorAdjust:"exact"},variants:[...Object.entries(e.palette).filter(b()).map(([r])=>({props:{color:r},style:{backgroundColor:I(e,r)}})),{props:({ownerState:r})=>r.color==="inherit"&&r.variant!=="buffer",style:{"&::before":{content:'""',position:"absolute",left:0,top:0,right:0,bottom:0,backgroundColor:"currentColor",opacity:.3}}},{props:{variant:"buffer"},style:{backgroundColor:"transparent"}},{props:{variant:"query"},style:{transform:"rotate(180deg)"}}]}))),E=y("span",{name:"MuiLinearProgress",slot:"Dashed",overridesResolver:(e,r)=>{const{ownerState:a}=e;return[r.dashed,r[`dashedColor${i(a.color)}`]]}})(C(({theme:e})=>({position:"absolute",marginTop:0,height:"100%",width:"100%",backgroundSize:"10px 10px",backgroundPosition:"0 -23px",variants:[{props:{color:"inherit"},style:{opacity:.3,backgroundImage:"radial-gradient(currentColor 0%, currentColor 16%, transparent 42%)"}},...Object.entries(e.palette).filter(b()).map(([r])=>{const a=I(e,r);return{props:{color:r},style:{backgroundImage:`radial-gradient(${a} 0%, ${a} 16%, transparent 42%)`}}})]})),z||{animation:`${$} 3s infinite linear`}),K=y("span",{name:"MuiLinearProgress",slot:"Bar1",overridesResolver:(e,r)=>{const{ownerState:a}=e;return[r.bar,r.bar1,r[`barColor${i(a.color)}`],(a.variant==="indeterminate"||a.variant==="query")&&r.bar1Indeterminate,a.variant==="determinate"&&r.bar1Determinate,a.variant==="buffer"&&r.bar1Buffer]}})(C(({theme:e})=>({width:"100%",position:"absolute",left:0,bottom:0,top:0,transition:"transform 0.2s linear",transformOrigin:"left",variants:[{props:{color:"inherit"},style:{backgroundColor:"currentColor"}},...Object.entries(e.palette).filter(b()).map(([r])=>({props:{color:r},style:{backgroundColor:(e.vars||e).palette[r].main}})),{props:{variant:"determinate"},style:{transition:`transform .${h}s linear`}},{props:{variant:"buffer"},style:{zIndex:1,transition:`transform .${h}s linear`}},{props:({ownerState:r})=>r.variant==="indeterminate"||r.variant==="query",style:{width:"auto"}},{props:({ownerState:r})=>r.variant==="indeterminate"||r.variant==="query",style:O||{animation:`${P} 2.1s cubic-bezier(0.65, 0.815, 0.735, 0.395) infinite`}}]}))),F=y("span",{name:"MuiLinearProgress",slot:"Bar2",overridesResolver:(e,r)=>{const{ownerState:a}=e;return[r.bar,r.bar2,r[`barColor${i(a.color)}`],(a.variant==="indeterminate"||a.variant==="query")&&r.bar2Indeterminate,a.variant==="buffer"&&r.bar2Buffer]}})(C(({theme:e})=>({width:"100%",position:"absolute",left:0,bottom:0,top:0,transition:"transform 0.2s linear",transformOrigin:"left",variants:[...Object.entries(e.palette).filter(b()).map(([r])=>({props:{color:r},style:{"--LinearProgressBar2-barColor":(e.vars||e).palette[r].main}})),{props:({ownerState:r})=>r.variant!=="buffer"&&r.color!=="inherit",style:{backgroundColor:"var(--LinearProgressBar2-barColor, currentColor)"}},{props:({ownerState:r})=>r.variant!=="buffer"&&r.color==="inherit",style:{backgroundColor:"currentColor"}},{props:{color:"inherit"},style:{opacity:.3}},...Object.entries(e.palette).filter(b()).map(([r])=>({props:{color:r,variant:"buffer"},style:{backgroundColor:I(e,r),transition:`transform .${h}s linear`}})),{props:({ownerState:r})=>r.variant==="indeterminate"||r.variant==="query",style:{width:"auto"}},{props:({ownerState:r})=>r.variant==="indeterminate"||r.variant==="query",style:S||{animation:`${x} 2.1s cubic-bezier(0.165, 0.84, 0.44, 1) 1.15s infinite`}}]}))),V=c.forwardRef(function(r,a){const t=q({props:r,name:"MuiLinearProgress"}),{className:u,color:s="primary",value:f,valueBuffer:d,variant:o="indeterminate",...B}=t,n={...t,color:s,variant:o},m=A(n),j=M(),g={},v={bar1:{},bar2:{}};if((o==="determinate"||o==="buffer")&&f!==void 0){g["aria-valuenow"]=Math.round(f),g["aria-valuemin"]=0,g["aria-valuemax"]=100;let l=f-100;j&&(l=-l),v.bar1.transform=`translateX(${l}%)`}if(o==="buffer"&&d!==void 0){let l=(d||0)-100;j&&(l=-l),v.bar2.transform=`translateX(${l}%)`}return p.jsxs(U,{className:T(m.root,u),ownerState:n,role:"progressbar",...g,ref:a,...B,children:[o==="buffer"?p.jsx(E,{className:m.dashed,ownerState:n}):null,p.jsx(K,{className:m.bar1,ownerState:n,style:v.bar1}),o==="determinate"?null:p.jsx(F,{className:m.bar2,ownerState:n,style:v.bar2})]})}),H=function(e){return p.jsx(V,{variant:"determinate",value:e.value,sx:{"& .MuiLinearProgress-determinate":{transition:"transform .2s linear"},height:e.height?e.height:r=>r.spacing(),marginTop:r=>r.spacing()}})};function _(e){const[a,t]=c.useState(void 0),[u,s]=c.useState(0),f=c.useCallback(()=>{s(0),t(new Date)},[t,s]),d=c.useCallback(()=>{s(0),t(void 0)},[]);return c.useEffect(()=>{if(!a)return;const o=setInterval(()=>{let n=(a?new Date().getTime()-a.getTime():0)/e*100;n>=100&&(n=100,t(void 0)),s(n)},100);return()=>clearInterval(o)},[a,s,t,e]),[u,f,d]}export{H as L,_ as u};
