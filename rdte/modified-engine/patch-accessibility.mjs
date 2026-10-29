#!/usr/bin/env node
// GPL engine additional semantic accessibility slice (no commercial source/data).
import {readFileSync,writeFileSync} from 'node:fs';
import {join} from 'node:path';
const root=process.argv[2];if(!root)throw Error('engine source root required');
const get=p=>readFileSync(join(root,p),'utf8');
const put=(p,s)=>writeFileSync(join(root,p),s);
let h=get('src/Engine/State.h');
if(!h.includes('#include <vector>\n')||!h.includes('void showAll();'))throw Error('State.h anchors differ');
h=h.replace('#include <vector>\n','#include <vector>\n#include <iosfwd>\n');
h=h.replace('void showAll();','void showAll();\n\tvoid sfRdteAccessibility(std::ostream& output) const;');
let cpp=get('src/Engine/State.cpp');
if(!cpp.includes('#include "State.h"\n')||!cpp.includes('namespace OpenXcom\n{\n'))throw Error('State.cpp anchors differ');
cpp=cpp.replace('#include "State.h"\n','#include "State.h"\n#include <ostream>\n#include "../Interface/Text.h"\n');
cpp=cpp.replace('namespace OpenXcom\n{\n','namespace OpenXcom\n{\n'+"\nstatic std::string sfAccessEscape(const std::string& text)\n{\n std::string escaped;\n for(unsigned char c:text)\n {\n  if(c=='\"'||c=='\\\\'){escaped.push_back('\\\\');escaped.push_back(char(c));}\n  else if(c<32)escaped.push_back(' ');\n  else escaped.push_back(char(c));\n }\n return escaped;\n}\nvoid State::sfRdteAccessibility(std::ostream& out) const\n{\n out<<\"[\";\n bool first=true;\n for(const Surface* surface:_surfaces)\n {\n  if(!surface||!surface->getVisible())continue;\n  const TextButton* button=dynamic_cast<const TextButton*>(surface);\n  const Text* text=dynamic_cast<const Text*>(surface);\n  if(!button&&!text)continue;\n  const std::string value=button?button->getText():text->getText();\n  if(value.empty())continue;\n  if(!first)out<<\",\";\n  first=false;\n  out<<\"{\\\"role\\\":\\\"\"<<(button?\"button\":\"text\")<<\"\\\",\\\"label\\\":\\\"\"<<sfAccessEscape(value)<<\"\\\",\\\"x\\\":\"\n     <<surface->getX()<<\",\\\"y\\\":\"<<surface->getY()<<\",\\\"w\\\":\"<<surface->getWidth()\n     <<\",\\\"h\\\":\"<<surface->getHeight()<<\"}\";\n }\n out<<\"]\";\n}\n");
let game=get('src/Engine/Game.cpp');
const start=game.indexOf('static void sfRdteState(const char* name)');
const end=game.indexOf('// SF_RDTE_EVAL_END',start);
if(start<0||end<0||!game.includes('sfRdteState(typeid(*_states.back()).name());'))throw Error('instrumented Game.cpp anchor missing');
game=game.slice(0,start)+"static void sfRdteState(const char* name,const State* state,double scaleX,double scaleY)\n{\n const char* path=std::getenv(\"OXCE_RDTE_TRACE\");\n if(!path||!*path||!name)return;\n static Uint32 lastTick=0;\n static std::string last;\n Uint32 now=SDL_GetTicks();\n if(now-lastTick<250)return;\n lastTick=now;\n std::ostringstream view;\n if(state)state->sfRdteAccessibility(view);\n else view<<\"[]\";\n const std::string value=std::string(name)+view.str();\n if(value==last)return;\n last=value;\n std::ofstream out(path,std::ios::app);\n if(out)out<<\"{\\\"event\\\":\\\"screen\\\",\\\"class\\\":\\\"\"<<name\n           <<\"\\\",\\\"scaleX\\\":\"<<scaleX<<\",\\\"scaleY\\\":\"<<scaleY\n           <<\",\\\"controls\\\":\"<<view.str()<<\"}\"<<std::endl;\n}\n"+game.slice(end);
game=game.replace('sfRdteState(typeid(*_states.back()).name());','sfRdteState(typeid(*_states.back()).name(),_states.back(),_screen->getXScale(),_screen->getYScale());');
put('src/Engine/State.h',h);put('src/Engine/State.cpp',cpp);put('src/Engine/Game.cpp',game);
console.log('SF_RDTE_SEMANTIC_ACCESSIBILITY_PATCH_APPLIED');
