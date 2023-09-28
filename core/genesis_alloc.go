// Copyright 2017 by the Authors
// This file is part of the go-core library.
//
// The go-core library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-core library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-core library. If not, see <http://www.gnu.org/licenses/>.

package core

// Constants containing the genesis allocation of built-in genesis blocks.
// Their content is an RLP-encoded list of (address, balance) tuples.
// Use mkalloc.go to create/update them.

// nolint: misspell
const mainnetAllocData = "\xf8\xfc\xe3\x96\xcb\x06+\r~\x13\xb4|@5\r+\xb7y@\bG7\u07aa\xb7U\x8bYU\xe3\xbb>t?\xec\x00\x00\x00\xe3\x96\xcb5}+*l\x1doAi\xf3\xb6\x18\xf9S\xea\x9e#q\xa9\u0632\x8b\x18\u043fB<\x03\xd8\xde\x00\x00\x00\xe3\x96\xcbUZ\xa3BQ\xab475\x9b\f\xa6_\xddU\xf6u\x85X\xac\xa1\x8b\x16Ux\xee\u03dd\x0f\xfb\x00\x00\x00\xe3\x96\u02c9\xe8In:\xab\x9bM\xee\x80\\\x92\xc5\u06c6\x057\x80\xc0\x13\xeb\x8b\x18\u043fB<\x03\xd8\xde\x00\x00\x00\xe3\x96\u02d3\x043h.\\\xd7&\xd9\xf6\x06\x9f\b\u017e/\xc6F\v\xaf\xf4\x8b\x04\xf6\x8c\xa6\xd8\u0351\xc6\x00\x00\x00\xe3\x96\u02d4\x85\xe8R=\xff\xd7P\x10,\xd0<\"\x87h\xe3\x00(\xd8\xf5\x03\x8b9\x13Q~\xbd<\fe\x00\x00\x00\xe3\x96\u02d5\x16\xeb\x8ae\xb7`\xd9\xd6&\xeb\xdc3\xc2\"\xfek^\x8bp\xe0\x8b\x18\u043fB<\x03\xd8\xde\x00\x00\x00"
const devinAllocData = "\xf9\f\xb3\u0756\xab\x02\x83\xb1\xc8W\xdc_\x80\x19\x1d2\U0008ddcb\xb5\u0539s\x97\\\x85\xe8\u0525\x10\x00\u0756\xab\x02\xd7\xd46\xd6<v\x98~56\x1c\xf2\xc4m\x12H\x9b\u052b\xf7\x85\xe8\u0525\x10\x00\u0716\xab\x03\xa9\"\xeen\x14\x9bm\xa6\xda\xe0\xf8\x05PxR\xcf3[\xc5\x05\x84;\x9a\xca\x00\u0756\xab\x03\xf7\xf4,\x06\xfbq\xee\xbaI\xaf\xaey\xcdx\x03\xd4\x04\u0603\x8d\x85\xe8\u0525\x10\x00\u0756\xab\x06\xfc\b\xa2\xb4\n%\x88?.\x91\xbc\xa3m\xb8\x9a=\u05d9\x89\xba\x85\xe8\u0525\x10\x00\u0756\xab\t\x93f\xd0p\x1b\xb8\xf5u\xe2\x81vP\x17\a\xc7\a\x98DQ&\x85\x17Hv\xe8\x00\u0756\xab\x12bG\xb7Y\xac\xa8p\x94\x9f\xabmK\u02a2_\xe23Cg#\x85\xe8\u0525\x10\x00\u0756\xab\x13\u0298\x01\xbd\xcd(\xec\xf6\x05\x14\x90ut\xba\xc4{\xb0+\xb4A\x85\xe8\u0525\x10\x00\u0756\xab\x14\xe6\x01\xe8\xc5@\x85\xbb\xf9\x9c$\xeb\x9cG\xe5W\xba\x97\x88pI\x85\xe8\u0525\x10\x00\u0756\xab\x14\xfb\x8a\xd7\xc1\te\xff\xb1d\x0elpX\x95f\xa7\x06\xc5|;\x85\xe8\u0525\x10\x00\u0756\xab\x16,X\xb5\xd2*H\f:\xc0\xbf\xec\\\xc3\xed\xf2\x99`e\xb8A\x85\xe8\u0525\x10\x00\u0756\xab\x17\xb2\xef3\xda\x02)\xff\x99t\x85-\x9b\br\xdax\x11\xfbR\x97\x85\xe8\u0525\x10\x00\u0756\xab!\x11\xe3N\xfe\xc0\x99g>\xa1%u\xaf\xe9]\x0f\x8bP\u0735\x1d\x85\xe8\u0525\x10\x00\u0756\xab!\x1d\x8ei\xf8^\xa3\x01\xc1Sy\xa3`\xb1%\x88\x12\xa6\xa0cV\x85\xe8\u0525\x10\x00\u0756\xab\"\u06d0O\x1b\xaf\xc1\xef,t\xe4j\xa9+3\x00\xeb\xd3\xeaq\xb5\x85\xe8\u0525\x10\x00\u0756\xab\"\xe18\xb7I\t\xc8\xe6\u0266,*~L\xd8&\u07f1\xcf\xd1J\x85\xe8\u0525\x10\x00\u0756\xab#\x8dn\xb5\x10D\x9a\xd4\xf3\x13C\xfbV\xd0p\xc0\xa1.\xea\x90\u0105\xe8\u0525\x10\x00\u0756\xab#\x98\xd20\xb5\xe10\xe2q\\\xfb\xa8\xf4\xf2\x96l\xc6\xca\n\xe8\u0745\xe8\u0525\x10\x00\u0716\xab$Cx\x13\xab\x89\b\x17\xd3M@\x97b\x95\xc6Z\x98o\xfb\xb1g\x84;\x9a\xca\x00\u0756\xab%\xb7J,=6\x98\x8d2_\x17\x9e\xc8\xe8G\x1c\xb3\xd6's\xee\x85\xe8\u0525\x10\x00\u0716\xab'\x18\x06\xb4S\x17\xb9\xb7\x87\xeb\xb3\x06k,\u03b6+\xc3}\xa2a\x84;\x9a\xca\x00\u0756\xab'A\u040c0\xfd\x87\x84\xd3k\x1b\x8c\x85`H\xe0\xfe\x02\u9c6e\x85\xe8\u0525\x10\x00\u0756\xab(Es\x17b\xfa_\x14\xc67F\xd4\xeb\xde\x1c\xd5\xe5j^\xb8\xfc\x85\xe8\u0525\x10\x00\u0756\xab0\x9d\x84j\x19\x19\xa2\xb0\xce\xfc\xc6+K>\x05+U\xf4Z\xeb^\x85\xe8\u0525\x10\x00\u0716\xab0\xc0q\u0728e\xac\xbc\xba\u0309\xf5\b\x84\xa8\xb9\x13\xce\x1e)r\x84;\x9a\xca\x00\u0716\xab0\xe7w\v\x83\xf4\x84\u06a7\xc47F\r\x84:T\xdcA\xd4l;\x84;\x9a\xca\x00\u0756\xab1\xd5P.Dr\xfa\x8b\xda9\x9f\x12o\x9fn\x95\xbb\xdc\u007f\xf1\u01c5\xe8\u0525\x10\x00\u0756\xab3\x19\xacR\xd0\u0525\xb2\x00\xdd\xca\xe4F{\xde\x04Fz\ub300\x85\xe8\u0525\x10\x00\u0756\xab4\xee\xb9\xd7'\x0e\xf9\t9\x18\xa2\xa4O@\xa7\xb3\xee\xc3/\xbe\u0185\xe8\u0525\x10\x00\u0756\xab69\x05x\x87k\x93\x95\xecv\x1aV\xd4\xce\x10\xab\x11\xab\xf5;\x0f\x85\xe8\u0525\x10\x00\u0716\xab69>\u02a2\xd3 \x9c\xee\x16\u039b#`\xe3'\xed<\x923F\x84;\x9a\xca\x00\u0716\xab6a\xdd\x14m\n\xf2\xd0\xc2[\xbc\xf8\xeb\xd8\xe1\xc5*\xa8j\x13h\x84;\x9a\xca\x00\u0756\xab6\xb0\x85\xb4\xa9\x11\xbb\v\"8\xaf-P\x01\bB3H\x18\xccD\x85\xe8\u0525\x10\x00\u0756\xab7L~\x86y\xe9\xcd%\x1fE\x1a0\xeb\xaao\xe4N\x8a\xae\xd9\xe9\x85\xe8\u0525\x10\x00\u0756\xab7l\x8b\a#:)>\b\x9e\x01\x8fSZ\x16\u07fa\x0e\xf5\x06.\x85\xe8\u0525\x10\x00\u0756\xab8sQ\xd2?\x86\xc4\x14\xe7\xc1\xa4\x109\xe7\u0758\xbd#\"_\xac\x85\xe8\u0525\x10\x00\u0716\xab9(\u00fe\x10\x86\x91\f\x00\xed\x0f\x1cL\x8f\x1a')\x15\xafO\x94\x84;\x9a\xca\x00\u0756\xab@\xf6-#\x0e\xea\xfaN\x00\x101\xb2O\x95\u06b1x\x0ff\x18\x18\x85\xe8\u0525\x10\x00\u0756\xabA\x88@\x9a\x12@\x03\xa9j\x80G%\x14\x98\u044a\xe2\x0e\xa6\x9de\x85\xe8\u0525\x10\x00\u0756\xabA\x97\xa4\x18Yku\x1bPX~sHH\x01\xf4\xfd\xc7\xccm\u0485\xe8\u0525\x10\x00\u0756\xabB\x9c\xc7<\xdbK\x81\xbf\x8d\xe1\r0j\x87\xf6P\xd8,@\xcaO\x85\xe8\u0525\x10\x00\u0756\xabCk\xfb:\xc2\xc0\xd9a$*\xb2Y\x1aX\xb0$C\xd6 \xc0?\x85\xe8\u0525\x10\x00\u0756\xabD\x82\xa2\xad\xceZ\x1d!\x13\xbcB\xdba[\x11\xe6\vE\u06b2z\x85\xe8\u0525\x10\x00\u0756\xabD\xdb\xdcA\x9c\xa2\x89}\x9b'\xc7\xfe\xa4v8\u052b\xf3\xa7\x84\x14\x85\xe8\u0525\x10\x00\u0756\xabD\xde5A>\xe2\xb6r\xd3\"\x93\x8e/\u0313-\\\f\xf8\uc205\xe8\u0525\x10\x00\u0756\xabD\xee\x04y0R\x1c\xf4\x91>\xe2\v\xc2T\xf1\xce\x18q!\xfd\x0e\x85\xe8\u0525\x10\x00\u0756\xabF\x10\xbe\xc9\xfbH\u0619[\x0eq2\xfb+\x9e\xde\xc4.5\xcc\xe0\x85\xe8\u0525\x10\x00\u0756\xabGC\x89G\xa4\xb8\xe1\xe9\x17nt\x1el\xc3\t[\xf9 F\x94G\x85\xe8\u0525\x10\x00\u0756\xabH\xe0X{Tm$\x98y\x97\xc9k\xb1\xafo\x85\xd4(8&\xa2\x85\xe8\u0525\x10\x00\u0716\xabH\xeb\x03G\x1a\x9b.\x8e\xd8\xe6\xd5J\u02b5\xb3P\xb7\xa1\xec \x0f\x84;\x9a\xca\x00\u0756\xabI\x90\u02ae'@a3\x95o\x934y\xa4DX\x9b\x11U\x17+\x85\xe8\u0525\x10\x00\u0716\xabP\x00\x9a;\xd5k\vB\xac\x88|a\xaf\x82J\u07ca\xb1\x84\xf25\x84;\x9a\xca\x00\u0716\xabP\x9bL\x1d\xa7[\xb5\xcb3g9\xef\x8e\xde\xcf\xe0\x94R\xbd\x87\b\x84;\x9a\xca\x00\u0756\xabRL\x81\xc9\u00a8S\xe6\xc4\xffkG\xb7\xd6c0\x8b\xab|V\x0f\x85\xe8\u0525\x10\x00\u0756\xabT\xacu\x91\x86\b\x02N\xdeRa?\xa1\xea\u0628\xa6\xb5\x97\xca\x17\x85\xe8\u0525\x10\x00\u0756\xabUGwmc\xe0'w\x86\x9fr\xcbR\xfa\xf6\xd9Ir$\xe4\u0685\xe8\u0525\x10\x00\u0756\xabU\xefv\xdf\xeb{$\x1f\x16^\xdcD\xa8<\x88\u007f\xec\x18\xd8\xc0o\x85\xe8\u0525\x10\x00\u0716\xabV\xd01Bn+NbS\a\u016e\xa5\xa4\xbd\x91qI\x1c\x1d'\x84;\x9a\xca\x00\u0756\xabW\xbbP|U\xa4\xae^\xeeF\xba\xbd0\xe6\r\x85p\xe1\xe3\xa3\u05c5\xe8\u0525\x10\x00\u0756\xabW\xdd\xe1\xa4pA\xfc<W\f\x03\x18\xa7\x13\x12\x8c\xedU\xfd*\u0685\xe8\u0525\x10\x00\u0756\xabX\vJ\xc7\xdb\xc7q\x9c0\xc5\fY\xe0\x8bi4\x97\xf1\u0210\xa4\x85\xe8\u0525\x10\x00\u0716\xabYyb\x10\xa3\xfe<$\xd43\x19z\xf0^\xf5L3'\x9b\xa8\r\x84;\x9a\xca\x00\u0756\xab`\x96Z\x94\x049l\x85\x17\xa5\x00\u0798/\xb2\xa1\xf2\xa5\x84W\u0685\xe8\u0525\x10\x00\u0756\xab`\xba\xa9N\xcb>\v\x0e\x04\x04a\u0450]\u0722{6\xbd\xad\xf5\x85\xe8\u0525\x10\x00\u0756\xabb]\x03\xbd(<\xa3R\x8dri\x1c\x1e\x12\u03a3\u008aI\xd0\x00\x85\xe8\u0525\x10\x00\u0756\xabb\xec\x01\xd7\x1c\xc1QN\x83\x87\"\u007f\xffC\xf0\u06c1\xd6\xf1\x9e\xe9\x85\xe8\u0525\x10\x00\u0756\xabb\xf0\xcf\r\xa5\xf0\u0382\xa6\xdd\x0eX\xe24\xf2\x8b\xe8\x88\xe6\x11\x05\x85\xe8\u0525\x10\x00\u0756\xabc\xf5\x0eG\xad\x00\x80\xc2\xeeE9\xf7\x89\xa3\u0282\\\xebE\xc4e\x85\xe8\u0525\x10\x00\u0756\xabe\xfc\xf4u\xd9\u012d\x89\xfbj\xd8h\xbc\U00085974/\xf1\xf2\xbe\x85\xe8\u0525\x10\x00\u0716\xabf\x0e\xf5\x11J\xd5:\x9f\xd1\x06\xb7*&\v\xa5\xb0U\xa9\xae\xca<\x84;\x9a\xca\x00\u0756\xabf8\x1a,\x17S\xca\u0501\x9c\x12\x80\x0f\x8d\xfa\x1fo\xc9T2\x82\x85\xe8\u0525\x10\x00\u0756\xabft\xe6\xfd\f|R\xb8\xee%\xb1\xabb\xe9\xe0\xe3\x85\nc\x92q\x85\xe8\u0525\x10\x00\u0756\xabf\x88!\xe8\"\x86\xfe\xd9\xec\x95\xc9\xcbE\t/\xcdBd\xfd\xba\u0105\xe8\u0525\x10\x00\u0756\xabhi\xce!\x0f\xe4$\xd9\u04d6rYb7\xba\x1c\x8fS:y\x89\x85\xe8\u0525\x10\x00\u0756\xabi\xc6;\xea\xd3\a\xfa\xa1\x1dY\x83l\x12A\x93\xe9\x8f|\xe57\xb1\x85\xe8\u0525\x10\x00\u0756\xabp\u078b\xaf\x8d\x84\u03f1_\xfdK\u0163\xcb\x19\x040gD*\x1f\x85\xe8\u0525\x10\x00\u0716\xabr3J>\xfbO\xb4\xaa\xb9\x04\u007fmQ\xf12\xf9\u4219\x1a\u0284;\x9a\xca\x00\u0756\xabr\xef\xfeN\v\x0e\x16\xaf73\x06d\xf0X\x06qV7W\x95\xfd\x85\xe8\u0525\x10\x00\u0756\xabs\x9e\x1d`(Y \x92\x8e\x88}\xb3\u017f\xbf\xa9\u0169]\xe2q\x85\xe8\u0525\x10\x00\u0756\xabu\xb2:\xb2V\x14\xd0\x1b~\u007f\xa3y\u01c0n:]\xe8\u03e6\xa5\x85\xe8\u0525\x10\x00\u0716\xaby\"\x15\xc4?\xc2\x13\xc0!\x82\xc88\x9f+\xc3$\b\xe2\xc5\t\"\x84;\x9a\xca\x00\u0756\xab\x80\x85\xa0\xd8\xe4\fh\xb1\xd6a4\u05cf#\xeb)\xfdT\xfeG\xaf\x85\xe8\u0525\x10\x00\u0756\xab\x80\u00ca\x9d\xc5\xd5\x01\xbc\xb0\xb8\"\x02\xe9Z\xe5\x8a\xfcWH\xb1C\x85\xe8\u0525\x10\x00\u0756\xab\x81\x1e@\x10\xdbDg*\x04\x16Q\u023d\xac\xbd,\xb5U\xc9n\xbf\x85\xe8\u0525\x10\x00\u0756\xab\x81E\xc4\v\xdf\xc3\xfc\n\xab\xab\xc1\r\x84BY\x88\xf0\x1c\xac[\x95\x85\xe8\u0525\x10\x00\u0756\xab\x81[f3^6\xc2\xe4)\x8f\xcb6\x9eb\xb4D[/\x86\x17D\x85\xe8\u0525\x10\x00\u0756\xab\x82\x1c>_\u0652\xde\x1eS\xddG\xa2.\xb6\x8fo%\xf2n\a\xa9\x85\xe8\u0525\x10\x00\u0716\xab\x82<\xe5=\x11\xed\n/iW\x0e\x9f{\n?\xac7\xc4s\x14\u0444;\x9a\xca\x00\u0756\xab\x82k\xf0\xfb9\xef\x03\x90\xfe\xec\x92\x1b\xbc\xdd\x1c\u0194\xf5W\x88 \x85\xe8\u0525\x10\x00\u0756\xab\x84}D\x1dwR\x1387\u03cf\xa4\xd0\xc2X\xb1r4\x9e\x02\xac\x85\xe8\u0525\x10\x00\u0756\xab\x84\x87\xb2\x1a\u007f\x11\xef\x9b^\xea\x97\x0f=\xc3PX\x15|\xb9|\xa1\x85\xe8\u0525\x10\x00\u0756\xab\x86{\xa3R0\xc4\xf0\x1f\u0394\xb0\xc5t\xbbCW\xa6\xca:k\xfd\x85\xe8\u0525\x10\x00\u0756\xab\x86\xf2\xdc\xe9\xd1(\xb7 \x81\x18o\xa1\xbf\xd3\xe6\xb6|\xcbQj\x87\x85\xe8\u0525\x10\x00\u0756\xab\x87 \xdb\xe0T\x15P\xf6\u07caWk\x0f\x92G`a\xffT\xd7\x00\x85\xe8\u0525\x10\x00\u0756\xab\x87r\x02'\xf6\u055cc!l\xff\xbe9[~/\xce\xf0%s\xb5\x85\xe8\u0525\x10\x00\u0756\xab\x87\x82gY5Ci\u007f\xf4\x01\x92m\xfb5\f\x97\x8a\x8e\x90\xe7a\x85\xe8\u0525\x10\x00\u0756\xab\x88\xd3\x1fY\x14\x18\xf0\u05c7\x9bY\xb7j\x05]v\xdd\xf8\xe8B\x03\x85\xe8\u0525\x10\x00\u0756\xab\x90\x1d\xf7O\xb4\x1d!\xb4W\v-,\x1e\xeb{\x18\xf9VU1\u0745\xe8\u0525\x10\x00\u0756\xab\x91\x88\xe1\xa1hQ\x16\xd3\xec\x19\xfe\xa3\xab\xc9\xf89h\xf9\xdf#\t\x85\xe8\u0525\x10\x00\u0756\xab\x91\x91\x81\xf7A\"\x13\u05f5`\x81\xd2\n)\xb0\x88\x17]\x10\th\x85\xe8\u0525\x10\x00\u0756\xab\x91\xf5\xc7\xcbY^N\xdf\xc3\x01\x06\x98\a*w\xf6\xc2z\x03>\xbe\x85\xe8\u0525\x10\x00\u0756\xab\x92-$\xe8\xfc\u0554#A\x1a\x0fT<\xa9J\xe72\x83\x87\x95$\x85\xe8\u0525\x10\x00\u0716\xab\x92XU\xd5\x15\x85C9\x06,n\x11\xd9\x14\xab\x14\x9a\v\x12\xafi\x84;\x9a\xca\x00\u0756\xab\x93\r~\xf0Y\xa7;z\xaf\xb7N~J\xbc\xc1o/\x05W<[\x85\xe8\u0525\x10\x00\u0756\xab\x97#i\u03d1\x88\x97\x17\xe0\xea\x95\xea\xc8k\xec8\xdc\u0217\x0f\u0685\xe8\u0525\x10\x00\u0756\xab\x98\x02m\x00\x94\x1e\xcbC\xe5S\x1fq\xc2I\xbe\x10P\xf9\xac\xb0w\x85\xe8\u0525\x10\x00\u0756\xab\x98\x17\xe7V\u007f@\xed\xda\x10\xd7\u0499\x17\r\x96@\x0e\x96ro\u0145\xe8\u0525\x10\x00\u0716\xab\x98\x19p\xd5kyr\xa7)\xee\xee\x1f&\xa3\xa9\xe0\x1d-\x14\x91\f\x84;\x9a\xca\x00\u0756\xab\x98\u01bd\xe9\xcd\xe5cl5\xb8\xad\x92\xbdv\x88oe\u0558tw\x85\xe8\u0525\x10\x00"
