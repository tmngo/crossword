function t(){}function n(t){return t()}function e(){return Object.create(null)}function o(t){t.forEach(n)}function r(t){return"function"==typeof t}function c(t,n){return t!=t?n==n:t!==n||t&&"object"==typeof t||"function"==typeof t}function a(t,n){t.appendChild(n)}function s(t,n,e){t.insertBefore(n,e||null)}function u(t){t.parentNode.removeChild(t)}function i(t,n){for(let e=0;e<t.length;e+=1)t[e]&&t[e].d(n)}function f(t){return document.createElement(t)}function l(t){return document.createTextNode(t)}function d(){return l(" ")}function p(t,n,e,o){return t.addEventListener(n,e,o),()=>t.removeEventListener(n,e,o)}function h(t){return function(n){return n.stopPropagation(),t.call(this,n)}}function $(t,n,e){null==e?t.removeAttribute(n):t.getAttribute(n)!==e&&t.setAttribute(n,e)}function m(t){return""===t?null:+t}function g(t,n){n=""+n,t.wholeText!==n&&(t.data=n)}function b(t,n){t.value=null==n?"":n}function y(t,n,e,o){t.style.setProperty(n,e,o?"important":"")}let _;function x(t){_=t}function v(){if(!_)throw new Error("Function called outside component initialization");return _}function E(t){v().$$.on_mount.push(t)}function w(){const t=v();return(n,e)=>{const o=t.$$.callbacks[n];if(o){const r=function(t,n,e=!1){const o=document.createEvent("CustomEvent");return o.initCustomEvent(t,e,!1,n),o}(n,e);o.slice().forEach((n=>{n.call(t,r)}))}}}const k=[],A=[],C=[],j=[],T=Promise.resolve();let N=!1;function O(t){C.push(t)}let P=!1;const S=new Set;function q(){if(!P){P=!0;do{for(let t=0;t<k.length;t+=1){const n=k[t];x(n),z(n.$$)}for(x(null),k.length=0;A.length;)A.pop()();for(let t=0;t<C.length;t+=1){const n=C[t];S.has(n)||(S.add(n),n())}C.length=0}while(k.length);for(;j.length;)j.pop()();N=!1,P=!1,S.clear()}}function z(t){if(null!==t.fragment){t.update(),o(t.before_update);const n=t.dirty;t.dirty=[-1],t.fragment&&t.fragment.p(t.ctx,n),t.after_update.forEach(O)}}const B=new Set;let L;function F(){L={r:0,c:[],p:L}}function M(){L.r||o(L.c),L=L.p}function D(t,n){t&&t.i&&(B.delete(t),t.i(n))}function G(t,n,e,o){if(t&&t.o){if(B.has(t))return;B.add(t),L.c.push((()=>{B.delete(t),o&&(e&&t.d(1),o())})),t.o(n)}}const H="undefined"!=typeof window?window:"undefined"!=typeof globalThis?globalThis:global;function I(t){t&&t.c()}function J(t,e,c,a){const{fragment:s,on_mount:u,on_destroy:i,after_update:f}=t.$$;s&&s.m(e,c),a||O((()=>{const e=u.map(n).filter(r);i?i.push(...e):o(e),t.$$.on_mount=[]})),f.forEach(O)}function K(t,n){const e=t.$$;null!==e.fragment&&(o(e.on_destroy),e.fragment&&e.fragment.d(n),e.on_destroy=e.fragment=null,e.ctx=[])}function Q(t,n){-1===t.$$.dirty[0]&&(k.push(t),N||(N=!0,T.then(q)),t.$$.dirty.fill(0)),t.$$.dirty[n/31|0]|=1<<n%31}function R(n,r,c,a,s,i,f,l=[-1]){const d=_;x(n);const p=n.$$={fragment:null,ctx:null,props:i,update:t,not_equal:s,bound:e(),on_mount:[],on_destroy:[],on_disconnect:[],before_update:[],after_update:[],context:new Map(d?d.$$.context:r.context||[]),callbacks:e(),dirty:l,skip_bound:!1,root:r.target||d.$$.root};f&&f(p.root);let h=!1;if(p.ctx=c?c(n,r.props||{},((t,e,...o)=>{const r=o.length?o[0]:e;return p.ctx&&s(p.ctx[t],p.ctx[t]=r)&&(!p.skip_bound&&p.bound[t]&&p.bound[t](r),h&&Q(n,t)),e})):[],p.update(),h=!0,o(p.before_update),p.fragment=!!a&&a(p.ctx),r.target){if(r.hydrate){const t=($=r.target,Array.from($.childNodes));p.fragment&&p.fragment.l(t),t.forEach(u)}else p.fragment&&p.fragment.c();r.intro&&D(n.$$.fragment),J(n,r.target,r.anchor,r.customElement),q()}var $;x(d)}class U{$destroy(){K(this,1),this.$destroy=t}$on(t,n){const e=this.$$.callbacks[t]||(this.$$.callbacks[t]=[]);return e.push(n),()=>{const t=e.indexOf(n);-1!==t&&e.splice(t,1)}}$set(t){var n;this.$$set&&(n=t,0!==Object.keys(n).length)&&(this.$$.skip_bound=!0,this.$$set(t),this.$$.skip_bound=!1)}}export{I as A,J as B,K as C,U as S,$ as a,s as b,w as c,u as d,f as e,A as f,a as g,g as h,R as i,d as j,y as k,p as l,b as m,t as n,E as o,h as p,G as q,o as r,c as s,l as t,M as u,D as v,m as w,i as x,H as y,F as z};
