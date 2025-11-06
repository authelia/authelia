import{v as R,w as S,r as c,x as q,b9 as D,j as f,y as g,z as M,O as n,D as N,K as v,Q as p,ba as L,bb as I,m as O}from"./index.BAKKlMLi.js";function T(e){return R("MuiLinearProgress",e)}S("MuiLinearProgress",["root","colorPrimary","colorSecondary","determinate","indeterminate","buffer","query","dashed","dashedColorPrimary","dashedColorSecondary","bar","bar1","bar2","barColorPrimary","barColorSecondary","bar1Indeterminate","bar1Determinate","bar1Buffer","bar2Indeterminate","bar2Buffer"]);const P=4,$=I`
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
`,w=typeof $!="string"?L`
        animation: ${$} 2.1s cubic-bezier(0.65, 0.815, 0.735, 0.395) infinite;
      `:null,k=I`
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
`,z=typeof k!="string"?L`
        animation: ${k} 2.1s cubic-bezier(0.165, 0.84, 0.44, 1) 1.15s infinite;
      `:null,x=I`
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
`,A=typeof x!="string"?L`
        animation: ${x} 3s infinite linear;
      `:null,U=e=>{const{classes:r,variant:a,color:t}=e,s={root:["root",`color${n(t)}`,a],dashed:["dashed",`dashedColor${n(t)}`],bar1:["bar","bar1",`barColor${n(t)}`,(a==="indeterminate"||a==="query")&&"bar1Indeterminate",a==="determinate"&&"bar1Determinate",a==="buffer"&&"bar1Buffer"],bar2:["bar","bar2",a!=="buffer"&&`barColor${n(t)}`,a==="buffer"&&`color${n(t)}`,(a==="indeterminate"||a==="query")&&"bar2Indeterminate",a==="buffer"&&"bar2Buffer"]};return N(s,T,r)},B=(e,r)=>e.vars?e.vars.palette.LinearProgress[`${r}Bg`]:e.palette.mode==="light"?e.lighten(e.palette[r].main,.62):e.darken(e.palette[r].main,.5),E=g("span",{name:"MuiLinearProgress",slot:"Root",overridesResolver:(e,r)=>{const{ownerState:a}=e;return[r.root,r[`color${n(a.color)}`],r[a.variant]]}})(v(({theme:e})=>({position:"relative",overflow:"hidden",display:"block",height:4,zIndex:0,"@media print":{colorAdjust:"exact"},variants:[...Object.entries(e.palette).filter(p()).map(([r])=>({props:{color:r},style:{backgroundColor:B(e,r)}})),{props:({ownerState:r})=>r.color==="inherit"&&r.variant!=="buffer",style:{"&::before":{content:'""',position:"absolute",left:0,top:0,right:0,bottom:0,backgroundColor:"currentColor",opacity:.3}}},{props:{variant:"buffer"},style:{backgroundColor:"transparent"}},{props:{variant:"query"},style:{transform:"rotate(180deg)"}}]}))),K=g("span",{name:"MuiLinearProgress",slot:"Dashed",overridesResolver:(e,r)=>{const{ownerState:a}=e;return[r.dashed,r[`dashedColor${n(a.color)}`]]}})(v(({theme:e})=>({position:"absolute",marginTop:0,height:"100%",width:"100%",backgroundSize:"10px 10px",backgroundPosition:"0 -23px",variants:[{props:{color:"inherit"},style:{opacity:.3,backgroundImage:"radial-gradient(currentColor 0%, currentColor 16%, transparent 42%)"}},...Object.entries(e.palette).filter(p()).map(([r])=>{const a=B(e,r);return{props:{color:r},style:{backgroundImage:`radial-gradient(${a} 0%, ${a} 16%, transparent 42%)`}}})]})),A||{animation:`${x} 3s infinite linear`}),V=g("span",{name:"MuiLinearProgress",slot:"Bar1",overridesResolver:(e,r)=>{const{ownerState:a}=e;return[r.bar,r.bar1,r[`barColor${n(a.color)}`],(a.variant==="indeterminate"||a.variant==="query")&&r.bar1Indeterminate,a.variant==="determinate"&&r.bar1Determinate,a.variant==="buffer"&&r.bar1Buffer]}})(v(({theme:e})=>({width:"100%",position:"absolute",left:0,bottom:0,top:0,transition:"transform 0.2s linear",transformOrigin:"left",variants:[{props:{color:"inherit"},style:{backgroundColor:"currentColor"}},...Object.entries(e.palette).filter(p()).map(([r])=>({props:{color:r},style:{backgroundColor:(e.vars||e).palette[r].main}})),{props:{variant:"determinate"},style:{transition:`transform .${P}s linear`}},{props:{variant:"buffer"},style:{zIndex:1,transition:`transform .${P}s linear`}},{props:({ownerState:r})=>r.variant==="indeterminate"||r.variant==="query",style:{width:"auto"}},{props:({ownerState:r})=>r.variant==="indeterminate"||r.variant==="query",style:w||{animation:`${$} 2.1s cubic-bezier(0.65, 0.815, 0.735, 0.395) infinite`}}]}))),X=g("span",{name:"MuiLinearProgress",slot:"Bar2",overridesResolver:(e,r)=>{const{ownerState:a}=e;return[r.bar,r.bar2,r[`barColor${n(a.color)}`],(a.variant==="indeterminate"||a.variant==="query")&&r.bar2Indeterminate,a.variant==="buffer"&&r.bar2Buffer]}})(v(({theme:e})=>({width:"100%",position:"absolute",left:0,bottom:0,top:0,transition:"transform 0.2s linear",transformOrigin:"left",variants:[...Object.entries(e.palette).filter(p()).map(([r])=>({props:{color:r},style:{"--LinearProgressBar2-barColor":(e.vars||e).palette[r].main}})),{props:({ownerState:r})=>r.variant!=="buffer"&&r.color!=="inherit",style:{backgroundColor:"var(--LinearProgressBar2-barColor, currentColor)"}},{props:({ownerState:r})=>r.variant!=="buffer"&&r.color==="inherit",style:{backgroundColor:"currentColor"}},{props:{color:"inherit"},style:{opacity:.3}},...Object.entries(e.palette).filter(p()).map(([r])=>({props:{color:r,variant:"buffer"},style:{backgroundColor:B(e,r),transition:`transform .${P}s linear`}})),{props:({ownerState:r})=>r.variant==="indeterminate"||r.variant==="query",style:{width:"auto"}},{props:({ownerState:r})=>r.variant==="indeterminate"||r.variant==="query",style:z||{animation:`${k} 2.1s cubic-bezier(0.165, 0.84, 0.44, 1) 1.15s infinite`}}]}))),_=c.forwardRef(function(r,a){const t=q({props:r,name:"MuiLinearProgress"}),{className:s,color:y="primary",value:u,valueBuffer:C,variant:o="indeterminate",...h}=t,i={...t,color:y,variant:o},b=U(i),j=D(),d={},m={bar1:{},bar2:{}};if((o==="determinate"||o==="buffer")&&u!==void 0){d["aria-valuenow"]=Math.round(u),d["aria-valuemin"]=0,d["aria-valuemax"]=100;let l=u-100;j&&(l=-l),m.bar1.transform=`translateX(${l}%)`}if(o==="buffer"&&C!==void 0){let l=(C||0)-100;j&&(l=-l),m.bar2.transform=`translateX(${l}%)`}return f.jsxs(E,{className:M(b.root,s),ownerState:i,role:"progressbar",...d,ref:a,...h,children:[o==="buffer"?f.jsx(K,{className:b.dashed,ownerState:i}):null,f.jsx(V,{className:b.bar1,ownerState:i,style:m.bar1}),o==="determinate"?null:f.jsx(X,{className:b.bar2,ownerState:i,style:m.bar2})]})}),H=function(e){const{classes:r}=F({props:e});return f.jsx(_,{variant:"determinate",classes:{root:r.root,determinate:r.determinate},value:e.value,className:r.default})},F=O()((e,{props:r})=>({root:{height:r.height?r.height:e.spacing()},determinate:{transition:"transform .2s linear"},default:{marginTop:e.spacing()}})),Q=100;function J(e){const[r,a]=c.useState(null),[t,s]=c.useState(0),y=c.useCallback(()=>{s(0),a(Date.now())},[]),u=c.useCallback(()=>{s(0),a(null)},[]);return c.useEffect(()=>{if(r===null)return;const o=setInterval(()=>{const h=Date.now()-r,i=Math.min(100,h/e*100);s(i),i>=100&&a(null)},Q);return()=>{clearInterval(o)}},[r,e]),[t,y,u]}export{H as L,J as u};
