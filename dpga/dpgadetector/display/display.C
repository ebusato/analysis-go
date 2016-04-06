
TEveBoxSet* display(TString fileName="../../godaq/dpgafullgeom.csv", Bool_t register=kTRUE)
{
  TEveManager::Create();

  gStyle->SetPalette(1, 0);
  
  TEveRGBAPalette* pal = new TEveRGBAPalette(0, 130);
  
  TEveBoxSet* q = new TEveBoxSet("BoxSet");
  q->SetPalette(pal);
  q->Reset(TEveBoxSet::kBT_FreeBox, kFALSE, 64);

  ifstream in(fileName.Data());
  if (!in) {
    cerr << "ERROR ! Unable to open file '" << fileName << "' !" << endl;
    return;
  }

  for(; !in.eof() ;) {
    string line;
    if (!getline(in,line)) break;
    if (!line.empty() && line[0]!='#') {
      istringstream istr(line);
      int iChannelAbs240;
      float X0, Y0, Z0;
      float X1, Y1, Z1;
      float X2, Y2, Z2;
      float X3, Y3, Z3;
      float X4, Y4, Z4;
      float X5, Y5, Z5;
      float X6, Y6, Z6;
      float X7, Y7, Z7;

      istr >> iChannelAbs240
	   >> X0 >> Y0 >> Z0
	   >> X1 >> Y1 >> Z1
	   >> X2 >> Y2 >> Z2
	   >> X3 >> Y3 >> Z3
	   >> X4 >> Y4 >> Z4
	   >> X5 >> Y5 >> Z5
	   >> X6 >> Y6 >> Z6
	   >> X7 >> Y7 >> Z7;
       
      Float_t verts[24] = {
	X0 , Y0 , Z0, 
	X1 , Y1 , Z1,
	X2 , Y2 , Z2,
	X3 , Y3 , Z3,
	X4 , Y4 , Z4,
	X5 , Y5 , Z5,
	X6 , Y6 , Z6,
	X7 , Y7 , Z7};
      q->AddBox(verts);
    }
  }
  q->RefitPlex();

  if (register)
    {
      gEve->AddElement(q);
      gEve->Redraw3D(kTRUE);
    }
  
  return q;
}
