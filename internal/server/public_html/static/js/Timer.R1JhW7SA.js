import{v as D,w as R,r as c,x as S,b9 as q,j as p,y,z as N,O as i,D as O,K as C,Q as b,ba as x,bb as L,m as w}from"./index.ChPl_EQa.js";function z(e){return D("MuiLinearProgress",e)}R("MuiLinearProgress",["root","colorPrimary","colorSecondary","determinate","indeterminate","buffer","query","dashed","dashedColorPrimary","dashedColorSecondary","bar","bar1","bar2","barColorPrimary","barColorSecondary","bar1Indeterminate","bar1Determinate","bar1Buffer","bar2Indeterminate","bar2Buffer"]);const h=4,P=L`
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
`,M=typeof P!="string"?x`
        animation: ${P} 2.1s cubic-bezier(0.65, 0.815, 0.735, 0.395) infinite;
      `:null,$=L`
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
`,T=typeof $!="string"?x`
        animation: ${$} 2.1s cubic-bezier(0.165, 0.84, 0.44, 1) 1.15s infinite;
      `:null,k=L`
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
`,A=typeof k!="string"?x`
        animation: ${k} 3s infinite linear;
      `:null,U=e=>{const{classes:r,variant:a,color:t}=e,u={root:["root",`color${i(t)}`,a],dashed:["dashed",`dashedColor${i(t)}`],bar1:["bar","bar1",`barColor${i(t)}`,(a==="indeterminate"||a==="query")&&"bar1Indeterminate",a==="determinate"&&"bar1Determinate",a==="buffer"&&"bar1Buffer"],bar2:["bar","bar2",a!=="buffer"&&`barColor${i(t)}`,a==="buffer"&&`color${i(t)}`,(a==="indeterminate"||a==="query")&&"bar2Indeterminate",a==="buffer"&&"bar2Buffer"]};return O(u,z,r)},I=(e,r)=>e.vars?e.vars.palette.LinearProgress[`${r}Bg`]:e.palette.mode==="light"?e.lighten(e.palette[r].main,.62):e.darken(e.palette[r].main,.5),K=y("span",{name:"MuiLinearProgress",slot:"Root",overridesResolver:(e,r)=>{const{ownerState:a}=e;return[r.root,r[`color${i(a.color)}`],r[a.variant]]}})(C(({theme:e})=>({position:"relative",overflow:"hidden",display:"block",height:4,zIndex:0,"@media print":{colorAdjust:"exact"},variants:[...Object.entries(e.palette).filter(b()).map(([r])=>({props:{color:r},style:{backgroundColor:I(e,r)}})),{props:({ownerState:r})=>r.color==="inherit"&&r.variant!=="buffer",style:{"&::before":{content:'""',position:"absolute",left:0,top:0,right:0,bottom:0,backgroundColor:"currentColor",opacity:.3}}},{props:{variant:"buffer"},style:{backgroundColor:"transparent"}},{props:{variant:"query"},style:{transform:"rotate(180deg)"}}]}))),E=y("span",{name:"MuiLinearProgress",slot:"Dashed",overridesResolver:(e,r)=>{const{ownerState:a}=e;return[r.dashed,r[`dashedColor${i(a.color)}`]]}})(C(({theme:e})=>({position:"absolute",marginTop:0,height:"100%",width:"100%",backgroundSize:"10px 10px",backgroundPosition:"0 -23px",variants:[{props:{color:"inherit"},style:{opacity:.3,backgroundImage:"radial-gradient(currentColor 0%, currentColor 16%, transparent 42%)"}},...Object.entries(e.palette).filter(b()).map(([r])=>{const a=I(e,r);return{props:{color:r},style:{backgroundImage:`radial-gradient(${a} 0%, ${a} 16%, transparent 42%)`}}})]})),A||{animation:`${k} 3s infinite linear`}),X=y("span",{name:"MuiLinearProgress",slot:"Bar1",overridesResolver:(e,r)=>{const{ownerState:a}=e;return[r.bar,r.bar1,r[`barColor${i(a.color)}`],(a.variant==="indeterminate"||a.variant==="query")&&r.bar1Indeterminate,a.variant==="determinate"&&r.bar1Determinate,a.variant==="buffer"&&r.bar1Buffer]}})(C(({theme:e})=>({width:"100%",position:"absolute",left:0,bottom:0,top:0,transition:"transform 0.2s linear",transformOrigin:"left",variants:[{props:{color:"inherit"},style:{backgroundColor:"currentColor"}},...Object.entries(e.palette).filter(b()).map(([r])=>({props:{color:r},style:{backgroundColor:(e.vars||e).palette[r].main}})),{props:{variant:"determinate"},style:{transition:`transform .${h}s linear`}},{props:{variant:"buffer"},style:{zIndex:1,transition:`transform .${h}s linear`}},{props:({ownerState:r})=>r.variant==="indeterminate"||r.variant==="query",style:{width:"auto"}},{props:({ownerState:r})=>r.variant==="indeterminate"||r.variant==="query",style:M||{animation:`${P} 2.1s cubic-bezier(0.65, 0.815, 0.735, 0.395) infinite`}}]}))),F=y("span",{name:"MuiLinearProgress",slot:"Bar2",overridesResolver:(e,r)=>{const{ownerState:a}=e;return[r.bar,r.bar2,r[`barColor${i(a.color)}`],(a.variant==="indeterminate"||a.variant==="query")&&r.bar2Indeterminate,a.variant==="buffer"&&r.bar2Buffer]}})(C(({theme:e})=>({width:"100%",position:"absolute",left:0,bottom:0,top:0,transition:"transform 0.2s linear",transformOrigin:"left",variants:[...Object.entries(e.palette).filter(b()).map(([r])=>({props:{color:r},style:{"--LinearProgressBar2-barColor":(e.vars||e).palette[r].main}})),{props:({ownerState:r})=>r.variant!=="buffer"&&r.color!=="inherit",style:{backgroundColor:"var(--LinearProgressBar2-barColor, currentColor)"}},{props:({ownerState:r})=>r.variant!=="buffer"&&r.color==="inherit",style:{backgroundColor:"currentColor"}},{props:{color:"inherit"},style:{opacity:.3}},...Object.entries(e.palette).filter(b()).map(([r])=>({props:{color:r,variant:"buffer"},style:{backgroundColor:I(e,r),transition:`transform .${h}s linear`}})),{props:({ownerState:r})=>r.variant==="indeterminate"||r.variant==="query",style:{width:"auto"}},{props:({ownerState:r})=>r.variant==="indeterminate"||r.variant==="query",style:T||{animation:`${$} 2.1s cubic-bezier(0.165, 0.84, 0.44, 1) 1.15s infinite`}}]}))),Q=c.forwardRef(function(r,a){const t=S({props:r,name:"MuiLinearProgress"}),{className:u,color:s="primary",value:f,valueBuffer:d,variant:o="indeterminate",...B}=t,n={...t,color:s,variant:o},m=U(n),j=q(),g={},v={bar1:{},bar2:{}};if((o==="determinate"||o==="buffer")&&f!==void 0){g["aria-valuenow"]=Math.round(f),g["aria-valuemin"]=0,g["aria-valuemax"]=100;let l=f-100;j&&(l=-l),v.bar1.transform=`translateX(${l}%)`}if(o==="buffer"&&d!==void 0){let l=(d||0)-100;j&&(l=-l),v.bar2.transform=`translateX(${l}%)`}return p.jsxs(K,{className:N(m.root,u),ownerState:n,role:"progressbar",...g,ref:a,...B,children:[o==="buffer"?p.jsx(E,{className:m.dashed,ownerState:n}):null,p.jsx(X,{className:m.bar1,ownerState:n,style:v.bar1}),o==="determinate"?null:p.jsx(F,{className:m.bar2,ownerState:n,style:v.bar2})]})}),G=function(e){const{classes:r}=V({props:e});return p.jsx(Q,{variant:"determinate",classes:{root:r.root,determinate:r.determinate},value:e.value,className:r.default})},V=w()((e,{props:r})=>({root:{height:r.height?r.height:e.spacing()},determinate:{transition:"transform .2s linear"},default:{marginTop:e.spacing()}}));function H(e){const[a,t]=c.useState(void 0),[u,s]=c.useState(0),f=c.useCallback(()=>{s(0),t(new Date)},[t,s]),d=c.useCallback(()=>{s(0),t(void 0)},[]);return c.useEffect(()=>{if(!a)return;const o=setInterval(()=>{let n=(a?new Date().getTime()-a.getTime():0)/e*100;n>=100&&(n=100,t(void 0)),s(n)},100);return()=>clearInterval(o)},[a,s,t,e]),[u,f,d]}export{G as L,H as u};
